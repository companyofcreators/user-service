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

// UpdateMasterProfileUseCase handles updating a master's profile fields.
type UpdateMasterProfileUseCase struct {
	masterRepo domain.MasterProfileRepository
	validator  *validator.Validate
	logger     *slog.Logger
}

func NewUpdateMasterProfileUseCase(
	masterRepo domain.MasterProfileRepository,
	logger *slog.Logger,
) *UpdateMasterProfileUseCase {
	return &UpdateMasterProfileUseCase{
		masterRepo: masterRepo,
		validator:  validator.New(),
		logger:     logger,
	}
}

func (u *UpdateMasterProfileUseCase) Execute(ctx context.Context, userID uuid.UUID, input domain.UpdateMasterProfileInput) (*domain.MasterProfile, error) {
	if err := u.validator.Struct(input); err != nil {
		u.logger.WarnContext(ctx, "validation failed for master profile update",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	existing, err := u.masterRepo.FindByUserID(ctx, userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to find master profile for update",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to find master profile: %w", err)
	}

	if existing == nil {
		return nil, fmt.Errorf("master profile not found")
	}

	if input.Description != nil {
		existing.Description = *input.Description
	}
	if input.ExperienceYears != nil {
		existing.ExperienceYears = *input.ExperienceYears
	}

	existing.UpdatedAt = time.Now().UTC()

	if err := u.masterRepo.Update(ctx, existing); err != nil {
		u.logger.ErrorContext(ctx, "failed to update master profile",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to update master profile: %w", err)
	}

	u.logger.InfoContext(ctx, "master profile updated",
		slog.String("user_id", userID.String()),
	)

	return existing, nil
}
