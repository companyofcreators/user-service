package user

import (
	"context"

	"github.com/google/uuid"
)

// UserService defines domain-level operations for user profiles.
type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*FullProfile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, updates UpdateProfileInput) (*UserProfile, error)
	GetMasterProfile(ctx context.Context, userID uuid.UUID) (*MasterProfile, error)
	UpdateMasterProfile(ctx context.Context, userID uuid.UUID, updates UpdateMasterProfileInput) (*MasterProfile, error)
	EnableMasterRole(ctx context.Context, userID uuid.UUID) (*MasterProfile, error)
	DisableMasterRole(ctx context.Context, userID uuid.UUID) error
	GetRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// UpdateProfileInput carries the fields that can be updated on a user profile.
type UpdateProfileInput struct {
	FirstName *string    `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string    `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	AvatarURL *string    `json:"avatar_url,omitempty" validate:"omitempty,url"`
	Phone     *string    `json:"phone,omitempty" validate:"omitempty,min=5,max=20"`
	Birthdate *string    `json:"birthdate,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// UpdateMasterProfileInput carries the fields that can be updated on a master profile.
type UpdateMasterProfileInput struct {
	Description     *string `json:"description,omitempty" validate:"omitempty,min=1,max=2000"`
	ExperienceYears *int    `json:"experience_years,omitempty" validate:"omitempty,min=0,max=100"`
}
