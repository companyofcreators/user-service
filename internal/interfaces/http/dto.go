package http

import (
	"time"

	"github.com/google/uuid"

	domain "github.com/companyofcreators/user-service/internal/domain/user"
)

// --- Request DTOs ---

// UpdateProfileRequest represents the PATCH body for updating a user profile.
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Patronymic *string `json:"patronymic,omitempty" validate:"omitempty,max=100"`
	AvatarURL  *string `json:"avatar_url,omitempty" `
	Phone     *string `json:"phone,omitempty" validate:"omitempty,min=5,max=20"`
	Birthdate *string `json:"birthdate,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

func (r UpdateProfileRequest) ToDomain() domain.UpdateProfileInput {
	return domain.UpdateProfileInput{
		FirstName:  r.FirstName,
		LastName:   r.LastName,
		Patronymic: r.Patronymic,
		AvatarURL:  r.AvatarURL,
		Phone:     r.Phone,
		Birthdate: r.Birthdate,
	}
}

// UpdateMasterProfileRequest represents the PATCH body for updating a master profile.
type UpdateMasterProfileRequest struct {
	Description     *string `json:"description,omitempty" validate:"omitempty,min=1,max=2000"`
	ExperienceYears *int    `json:"experience_years,omitempty" validate:"omitempty,min=0,max=100"`
}

func (r UpdateMasterProfileRequest) ToDomain() domain.UpdateMasterProfileInput {
	return domain.UpdateMasterProfileInput{
		Description:     r.Description,
		ExperienceYears: r.ExperienceYears,
	}
}

// --- Response DTOs ---

// ErrorResponse is a standard error JSON body.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}

// PublicProfileResponse is the JSON response for a public user profile.
type PublicProfileResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL string    `json:"avatar_url"`
}

// UserProfileResponse is the JSON response for a user profile.
type UserProfileResponse struct {
	ID        uuid.UUID  `json:"id"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Patronymic string     `json:"patronymic,omitempty"`
	AvatarURL  string     `json:"avatar_url"`
	Phone     string     `json:"phone"`
	Birthdate *time.Time `json:"birthdate,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func NewUserProfileResponse(p *domain.UserProfile) UserProfileResponse {
	return UserProfileResponse{
		ID:        p.ID,
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Patronymic: p.Patronymic,
		AvatarURL:  p.AvatarURL,
		Phone:     p.Phone,
		Birthdate: p.Birthdate,
		UpdatedAt: p.UpdatedAt,
	}
}

// MasterProfileResponse is the JSON response for a master profile.
type MasterProfileResponse struct {
	UserID          uuid.UUID `json:"user_id"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	AvatarURL       string    `json:"avatar_url"`
	IsActive        bool      `json:"is_active"`
	Description     string    `json:"description"`
	ExperienceYears int       `json:"experience_years"`
	Rating          float64   `json:"rating"`
	CompletedOrders int       `json:"completed_orders"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func NewMasterProfileResponse(mp *domain.MasterProfile, up *domain.UserProfile) MasterProfileResponse {
	resp := MasterProfileResponse{
		UserID:          mp.UserID,
		IsActive:        mp.IsActive,
		Description:     mp.Description,
		ExperienceYears: mp.ExperienceYears,
		Rating:          mp.Rating,
		CompletedOrders: mp.CompletedOrders,
		UpdatedAt:       mp.UpdatedAt,
	}
	if up != nil {
		resp.FirstName = up.FirstName
		resp.LastName = up.LastName
		resp.AvatarURL = up.AvatarURL
	}
	return resp
}

// FullProfileResponse is the JSON response for a full user profile.
type FullProfileResponse struct {
	Profile       *UserProfileResponse   `json:"profile,omitempty"`
	MasterProfile *MasterProfileResponse `json:"master_profile,omitempty"`
	Roles         []string               `json:"roles"`
}

func NewFullProfileResponse(fp *domain.FullProfile) FullProfileResponse {
	resp := FullProfileResponse{
		Roles: fp.Roles,
	}
	if fp.Profile != nil {
		pr := NewUserProfileResponse(fp.Profile)
		resp.Profile = &pr
	}
	if fp.MasterProfile != nil {
		mpr := NewMasterProfileResponse(fp.MasterProfile, fp.Profile)
		resp.MasterProfile = &mpr
	}
	return resp
}
