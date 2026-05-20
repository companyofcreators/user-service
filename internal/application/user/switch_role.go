package user

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/user-service/internal/domain/user"
)

// OrderChecker is an interface for checking if a user has active orders.
// Implemented by app.OrderClient (in package app) to avoid circular imports.
type OrderChecker interface {
	HasActiveOrders(ctx context.Context, userID string) (bool, error)
}

// SwitchRoleUseCase handles enabling or disabling the master role for a user.
type SwitchRoleUseCase struct {
	userRepo     domain.UserProfileRepository
	masterRepo   domain.MasterProfileRepository
	roleRepo     domain.UserRoleRepository
	orderChecker OrderChecker
	logger       *slog.Logger
}

func NewSwitchRoleUseCase(
	userRepo domain.UserProfileRepository,
	masterRepo domain.MasterProfileRepository,
	roleRepo domain.UserRoleRepository,
	orderChecker OrderChecker,
	logger *slog.Logger,
) *SwitchRoleUseCase {
	return &SwitchRoleUseCase{
		userRepo:     userRepo,
		masterRepo:   masterRepo,
		roleRepo:     roleRepo,
		orderChecker: orderChecker,
		logger:       logger,
	}
}

// EnableMasterRole adds the "master" role and creates a MasterProfile if one does not exist.
func (u *SwitchRoleUseCase) EnableMasterRole(ctx context.Context, userID uuid.UUID) (*domain.MasterProfile, error) {
	// Prevent self-dealing: cannot enable master while you have active orders as customer
	if u.orderChecker != nil {
		hasActive, err := u.orderChecker.HasActiveOrders(ctx, userID.String())
		if err != nil {
			u.logger.WarnContext(ctx, "failed to check active orders before enabling master role, proceeding anyway",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
		} else if hasActive {
			return nil, fmt.Errorf("cannot enable master while you have active orders as customer")
		}
	}

	// Ensure user profile exists
	userProfile, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to find user profile for role switch",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to find user profile: %w", err)
	}
	if userProfile == nil {
		return nil, fmt.Errorf("user profile not found")
	}

	// Add master role (idempotent - may already exist)
	if err := u.roleRepo.AddRole(ctx, userID, "master"); err != nil {
		u.logger.ErrorContext(ctx, "failed to add master role",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to add master role: %w", err)
	}

	// Create or reactivate master profile
	masterProfile, err := u.masterRepo.FindByUserID(ctx, userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to check master profile existence",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to check master profile: %w", err)
	}

	if masterProfile == nil {
		// Create new master profile
		masterProfile = &domain.MasterProfile{
			UserID:   userID,
			IsActive: true,
			UpdatedAt: time.Now().UTC(),
		}
		if err := u.masterRepo.Create(ctx, masterProfile); err != nil {
			u.logger.ErrorContext(ctx, "failed to create master profile",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to create master profile: %w", err)
		}
	} else {
		// Reactivate existing master profile
		masterProfile.IsActive = true
		masterProfile.UpdatedAt = time.Now().UTC()
		if err := u.masterRepo.Update(ctx, masterProfile); err != nil {
			u.logger.ErrorContext(ctx, "failed to reactivate master profile",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to reactivate master profile: %w", err)
		}
	}

	u.logger.InfoContext(ctx, "master role enabled",
		slog.String("user_id", userID.String()),
	)

	return masterProfile, nil
}

// DisableMasterRole removes the "master" role and sets the master profile as inactive.
func (u *SwitchRoleUseCase) DisableMasterRole(ctx context.Context, userID uuid.UUID) error {
	// Prevent self-dealing: cannot disable master while you have active orders in progress as master
	if u.orderChecker != nil {
		hasActive, err := u.orderChecker.HasActiveOrders(ctx, userID.String())
		if err != nil {
			u.logger.WarnContext(ctx, "failed to check active orders before disabling master role, proceeding anyway",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
		} else if hasActive {
			return fmt.Errorf("cannot disable master while you have active orders in progress")
		}
	}

	// Remove master role
	if err := u.roleRepo.RemoveRole(ctx, userID, "master"); err != nil {
		u.logger.ErrorContext(ctx, "failed to remove master role",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to remove master role: %w", err)
	}

	// Deactivate master profile
	masterProfile, err := u.masterRepo.FindByUserID(ctx, userID)
	if err != nil {
		u.logger.ErrorContext(ctx, "failed to find master profile for deactivation",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to find master profile: %w", err)
	}

	if masterProfile != nil {
		masterProfile.IsActive = false
		masterProfile.UpdatedAt = time.Now().UTC()
		if err := u.masterRepo.Update(ctx, masterProfile); err != nil {
			u.logger.ErrorContext(ctx, "failed to deactivate master profile",
				slog.String("user_id", userID.String()),
				slog.String("error", err.Error()),
			)
			return fmt.Errorf("failed to deactivate master profile: %w", err)
		}
	}

	u.logger.InfoContext(ctx, "master role disabled",
		slog.String("user_id", userID.String()),
	)

	return nil
}
