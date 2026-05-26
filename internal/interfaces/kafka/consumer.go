package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	domain "github.com/companyofcreators/user-service/internal/domain/user"
)

// UserCreatedEvent is the message received when a new user is registered.
type UserCreatedEvent struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}

// ReviewCreatedEvent is the message received when a review is submitted.
type ReviewCreatedEvent struct {
	ReviewID       string  `json:"review_id"`
	MasterID       string  `json:"master_id"`
	ReviewedUserID string  `json:"reviewed_user_id"`
	ToUserID       string  `json:"to_user_id"`
	Rating         float64 `json:"rating"`
	OrderID        string  `json:"order_id"`
}

// KafkaConsumer manages Kafka topic consumption for the user-service.
type KafkaConsumer struct {
	brokers     []string
	userRepo    domain.UserProfileRepository
	roleRepo    domain.UserRoleRepository
	masterRepo  domain.MasterProfileRepository
	logger      *slog.Logger
	cancelFuncs []context.CancelFunc
	mu          sync.Mutex
}

// NewKafkaConsumer creates a new KafkaConsumer.
func NewKafkaConsumer(
	brokers []string,
	userRepo domain.UserProfileRepository,
	roleRepo domain.UserRoleRepository,
	masterRepo domain.MasterProfileRepository,
	logger *slog.Logger,
) *KafkaConsumer {
	return &KafkaConsumer{
		brokers:    brokers,
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		masterRepo: masterRepo,
		logger:     logger,
	}
}

// Start starts all Kafka consumers in goroutines.
func (c *KafkaConsumer) Start(ctx context.Context) error {
	consumerCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	c.cancelFuncs = append(c.cancelFuncs, cancel)
	c.mu.Unlock()

	// Start user.created consumer
	go c.consumeUserCreated(consumerCtx)

	// Start review.created consumer
	go c.consumeReviewCreated(consumerCtx)

	c.logger.Info("kafka consumers started",
		slog.String("brokers", strings.Join(c.brokers, ",")),
	)

	return nil
}

// Shutdown gracefully stops all Kafka consumers.
func (c *KafkaConsumer) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cancel := range c.cancelFuncs {
		cancel()
	}
	c.logger.Info("kafka consumers shut down")
}

// consumeUserCreated listens for user.created events and creates empty profiles.
func (c *KafkaConsumer) consumeUserCreated(ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.brokers,
		Topic:       "user.created",
		GroupID:     "user-service-profile-creator",
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	c.logger.Info("consumer started", slog.String("topic", "user.created"))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer stopping", slog.String("topic", "user.created"))
			return
		default:
		}

		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("failed to read kafka message",
				slog.String("topic", "user.created"),
				slog.String("error", err.Error()),
			)
			continue
		}

		c.logger.Info("received message",
			slog.String("topic", "user.created"),
			slog.Int64("offset", msg.Offset),
			slog.Int("partition", msg.Partition),
		)

		var event UserCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal user.created event",
				slog.String("error", err.Error()),
			)
			continue
		}

		userID, err := uuid.Parse(event.UserID)
		if err != nil {
			c.logger.Error("invalid user_id in user.created event",
				slog.String("user_id", event.UserID),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Create empty user profile
		profile := &domain.UserProfile{
			ID:        userID,
			UpdatedAt: time.Now().UTC(),
		}

		if err := c.userRepo.Create(ctx, profile); err != nil {
			c.logger.Error("failed to create user profile from event",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Assign roles from the event
		roles := event.Roles
		if len(roles) == 0 {
			roles = []string{"user"}
		}
		for _, role := range roles {
			if err := c.roleRepo.AddRole(ctx, userID, role); err != nil {
				c.logger.Error("failed to add role from event",
					slog.String("user_id", userID.String()),
					slog.String("role", role),
					slog.String("error", err.Error()),
				)
				// Continue even if role assignment fails
			}
		}

		c.logger.Info("user profile created from event",
			slog.String("user_id", userID.String()),
			slog.String("email", event.Email),
		)
	}
}

// consumeReviewCreated listens for review.created events and recalculates master rating.
func (c *KafkaConsumer) consumeReviewCreated(ctx context.Context) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     c.brokers,
		Topic:       "review.created",
		GroupID:     "user-service-rating-calculator",
		StartOffset: kafka.FirstOffset,
	})
	defer reader.Close()

	c.logger.Info("consumer started", slog.String("topic", "review.created"))

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer stopping", slog.String("topic", "review.created"))
			return
		default:
		}

		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.logger.Error("failed to read kafka message",
				slog.String("topic", "review.created"),
				slog.String("error", err.Error()),
			)
			continue
		}

		c.logger.Info("received message",
			slog.String("topic", "review.created"),
			slog.Int64("offset", msg.Offset),
			slog.Int("partition", msg.Partition),
		)

		var event ReviewCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			c.logger.Error("failed to unmarshal review.created event",
				slog.String("error", err.Error()),
			)
			continue
		}

		if event.MasterID == "" {
			event.MasterID = event.ReviewedUserID
		}
		if event.MasterID == "" {
			event.MasterID = event.ToUserID
		}
		masterID, err := uuid.Parse(event.MasterID)
		if err != nil {
			c.logger.Error("invalid master_id in review.created event",
				slog.String("master_id", event.MasterID),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Update the master's rating with the new review rating
		// In a production system, this would recalculate the average from all reviews.
		// For now, use a rolling average approach with the master's current completed_orders count.
		masterProfile, err := c.masterRepo.FindByUserID(ctx, masterID)
		if err != nil {
			c.logger.Error("failed to find master profile for rating update",
				slog.String("master_id", masterID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		if masterProfile == nil {
			c.logger.Warn("master profile not found for rating update, skipping",
				slog.String("master_id", masterID.String()),
			)
			continue
		}

		// Calculate new rolling average rating
		// new_avg = (old_avg * count + new_rating) / (count + 1)
		newRating := event.Rating
		if masterProfile.CompletedOrders > 0 {
			newRating = (masterProfile.Rating*float64(masterProfile.CompletedOrders) + event.Rating) / float64(masterProfile.CompletedOrders+1)
		}

		if err := c.masterRepo.UpdateRating(ctx, masterID, newRating); err != nil {
			c.logger.Error("failed to update master rating",
				slog.String("master_id", masterID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		// Increment completed orders count
		if err := c.masterRepo.IncrementCompletedOrders(ctx, masterID); err != nil {
			c.logger.Error("failed to increment completed orders",
				slog.String("master_id", masterID.String()),
				slog.String("error", err.Error()),
			)
			continue
		}

		c.logger.Info("master rating updated from review event",
			slog.String("review_id", event.ReviewID),
			slog.String("master_id", masterID.String()),
			slog.Float64("new_rating", newRating),
		)
	}
}
