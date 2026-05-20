package user

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/user-service/internal/domain/user"
)

// GetProfileUseCase retrieves a user's full profile, including master profile and roles if they exist.
type GetProfileUseCase struct {
	userRepo   domain.UserProfileRepository
	masterRepo domain.MasterProfileRepository
	roleRepo   domain.UserRoleRepository
	logger     *slog.Logger
}

func NewGetProfileUseCase(
	userRepo domain.UserProfileRepository,
	masterRepo domain.MasterProfileRepository,
	roleRepo domain.UserRoleRepository,
	logger *slog.Logger,
) *GetProfileUseCase {
	return &GetProfileUseCase{
		userRepo:   userRepo,
		masterRepo: masterRepo,
		roleRepo:   roleRepo,
		logger:     logger,
	}
}

func (u *GetProfileUseCase) Execute(ctx context.Context, userID uuid.UUID) (*domain.FullProfile, error) {
	profile, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to find user profile", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to find user profile: %w", err)
	}

	if profile == nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	roles, err := u.roleRepo.GetRoles(ctx, userID)
	if err != nil {
		u.logger.WarnContext(ctx, "failed to get user roles", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
		roles = []string{}
	}

	result := &domain.FullProfile{
		Profile: profile,
		Roles:   roles,
	}

	// Master profile is optional - not all users are masters
	masterProfile, err := u.masterRepo.FindByUserID(ctx, userID)
	if err != nil {
		u.logger.WarnContext(ctx, "failed to get master profile", slog.String("user_id", userID.String()), slog.String("error", err.Error()))
	} else if masterProfile != nil {
		result.MasterProfile = masterProfile
	}

	return result, nil
}
