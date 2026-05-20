package app

import (
	"log/slog"

	"github.com/jmoiron/sqlx"

	appuser "github.com/companyofcreators/user-service/internal/application/user"
	domain "github.com/companyofcreators/user-service/internal/domain/user"
	"github.com/companyofcreators/user-service/internal/infrastructure/db"
	kafkaconsumer "github.com/companyofcreators/user-service/internal/interfaces/kafka"
)

// Container holds all wired dependencies for the user-service.
type Container struct {
	DB     *sqlx.DB
	Logger *slog.Logger

	// Repositories
	UserProfileRepo   domain.UserProfileRepository
	MasterProfileRepo domain.MasterProfileRepository
	UserRoleRepo      domain.UserRoleRepository

	// Use cases
	GetProfileUseCase          *appuser.GetProfileUseCase
	UpdateProfileUseCase       *appuser.UpdateProfileUseCase
	GetMasterProfileUseCase    *appuser.GetMasterProfileUseCase
	UpdateMasterProfileUseCase *appuser.UpdateMasterProfileUseCase
	SwitchRoleUseCase          *appuser.SwitchRoleUseCase

	// Infrastructure
	KafkaConsumer *kafkaconsumer.KafkaConsumer
}

// NewContainer creates and wires all dependencies.
func NewContainer(database *sqlx.DB, logger *slog.Logger, kafkaBrokers []string, orderServiceURL string) *Container {
	c := &Container{
		DB:     database,
		Logger: logger,
	}

	// Repositories
	c.UserProfileRepo = db.NewUserProfileRepo(database, logger)
	c.MasterProfileRepo = db.NewMasterProfileRepo(database, logger)
	c.UserRoleRepo = db.NewUserRoleRepo(database, logger)

	// Clients
	orderClient := NewOrderClient(orderServiceURL)

	// Use cases
	c.GetProfileUseCase = appuser.NewGetProfileUseCase(
		c.UserProfileRepo,
		c.MasterProfileRepo,
		c.UserRoleRepo,
		logger,
	)
	c.UpdateProfileUseCase = appuser.NewUpdateProfileUseCase(
		c.UserProfileRepo,
		logger,
	)
	c.GetMasterProfileUseCase = appuser.NewGetMasterProfileUseCase(
		c.MasterProfileRepo,
		logger,
	)
	c.UpdateMasterProfileUseCase = appuser.NewUpdateMasterProfileUseCase(
		c.MasterProfileRepo,
		logger,
	)
	c.SwitchRoleUseCase = appuser.NewSwitchRoleUseCase(
		c.UserProfileRepo,
		c.MasterProfileRepo,
		c.UserRoleRepo,
		orderClient,
		logger,
	)

	// Kafka consumer
	c.KafkaConsumer = kafkaconsumer.NewKafkaConsumer(
		kafkaBrokers,
		c.UserProfileRepo,
		c.UserRoleRepo,
		c.MasterProfileRepo,
		logger,
	)

	return c
}
