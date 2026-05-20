package user

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	domain "github.com/companyofcreators/user-service/internal/domain/user"
)

// UpdateProfileUseCase handles updating a user's profile fields.
type UpdateProfileUseCase struct {
	userRepo  domain.UserProfileRepository
	validator *validator.Validate
	logger    *slog.Logger
}

func NewUpdateProfileUseCase(
	userRepo domain.UserProfileRepository,
	logger *slog.Logger,
) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		userRepo:  userRepo,
		validator: validator.New(),
		logger:    logger,
	}
}

func (u *UpdateProfileUseCase) Execute(ctx context.Context, userID uuid.UUID, input domain.UpdateProfileInput) (*domain.UserProfile, error) {
	if err := u.validator.Struct(input); err != nil {
		u.logger.WarnContext(ctx, "validation failed for profile update",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	// Fetch existing profile
	existing, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to find user profile for update",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to find user profile: %w", err)
	}

	if existing == nil {
		return nil, fmt.Errorf("user profile not found")
	}

	// Apply updates only for non-nil fields
	if input.FirstName != nil {
		existing.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		existing.LastName = *input.LastName
	}
	if input.AvatarURL != nil {
		existing.AvatarURL = *input.AvatarURL
	}
	if input.Phone != nil {
		existing.Phone = *input.Phone
	}
	if input.Birthdate != nil {
		parsed, err := time.Parse("2006-01-02", *input.Birthdate)
		if err != nil {
			return nil, fmt.Errorf("invalid birthdate format: %w", err)
		}
		existing.Birthdate = &parsed
	}

	existing.UpdatedAt = time.Now().UTC()

	if err := u.userRepo.Update(ctx, existing); err != nil {
		u.logger.ErrorContext(ctx, "failed to update user profile",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	u.logger.InfoContext(ctx, "user profile updated",
		slog.String("user_id", userID.String()),
	)

	return existing, nil
}
