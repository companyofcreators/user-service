package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/user-service/internal/domain/user"
)

// GetMasterProfileUseCase retrieves a master profile by user ID.
type GetMasterProfileUseCase struct {
	masterRepo domain.MasterProfileRepository
	logger     *slog.Logger
}

func NewGetMasterProfileUseCase(
	masterRepo domain.MasterProfileRepository,
	logger *slog.Logger,
) *GetMasterProfileUseCase {
	return &GetMasterProfileUseCase{
		masterRepo: masterRepo,
		logger:     logger,
	}
}

func (u *GetMasterProfileUseCase) Execute(ctx context.Context, userID uuid.UUID) (*domain.MasterProfile, error) {
	profile, err := u.masterRepo.FindByUserID(ctx, userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to find master profile",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to find master profile: %w", err)
	}

	if profile == nil {
		return nil, fmt.Errorf("профиль мастера не найден")
	}

	return profile, nil
}
