package user

import (
	"context"

	"github.com/google/uuid"
)

// UserProfileRepository defines persistence operations for user profiles.
type UserProfileRepository interface {
	Create(ctx context.Context, profile *UserProfile) error
	FindByID(ctx context.Context, id uuid.UUID) (*UserProfile, error)
	Update(ctx context.Context, profile *UserProfile) error
}

// MasterProfileRepository defines persistence operations for master profiles.
type MasterProfileRepository interface {
	Create(ctx context.Context, mp *MasterProfile) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*MasterProfile, error)
	Update(ctx context.Context, mp *MasterProfile) error
	FindActive(ctx context.Context) ([]*MasterProfile, error)
	UpdateRating(ctx context.Context, userID uuid.UUID, rating float64) error
	IncrementCompletedOrders(ctx context.Context, userID uuid.UUID) error
}

// UserRoleRepository defines persistence operations for user roles.
type UserRoleRepository interface {
	GetRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	AddRole(ctx context.Context, userID uuid.UUID, role string) error
	RemoveRole(ctx context.Context, userID uuid.UUID, role string) error
}
