package user

import (
	"time"

	"github.com/google/uuid"
)

// UserProfile represents a public user profile.
// This is NOT for authentication - it only stores public-facing information.
type UserProfile struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	FirstName string     `json:"first_name" db:"first_name"`
	LastName  string     `json:"last_name" db:"last_name"`
	AvatarURL string     `json:"avatar_url" db:"avatar_url"`
	Phone     string     `json:"phone" db:"phone"`
	Birthdate *time.Time `json:"birthdate" db:"birthdate"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// MasterProfile represents a master's public profile with rating and statistics.
type MasterProfile struct {
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	IsActive        bool      `json:"is_active" db:"is_active"`
	Description     string    `json:"description" db:"description"`
	ExperienceYears int       `json:"experience_years" db:"experience_years"`
	Rating          float64   `json:"rating" db:"rating"`
	CompletedOrders int       `json:"completed_orders" db:"completed_orders"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents a single role assignment for a user.
type UserRole struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Role      string    `json:"role" db:"role"` // "user", "master", "moderator", "admin"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FullProfile aggregates all profile data for a user.
type FullProfile struct {
	Profile       *UserProfile   `json:"profile,omitempty"`
	MasterProfile *MasterProfile `json:"master_profile,omitempty"`
	Roles         []string       `json:"roles"`
}
