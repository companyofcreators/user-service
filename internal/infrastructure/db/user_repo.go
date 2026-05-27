package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/companyofcreators/user-service/internal/domain/user"
)

// Ensure interfaces are satisfied at compile time.
var (
	_ domain.UserProfileRepository   = (*UserProfileRepo)(nil)
	_ domain.MasterProfileRepository = (*MasterProfileRepo)(nil)
	_ domain.UserRoleRepository      = (*UserRoleRepo)(nil)
)

// --- UserProfileRepo ---

type UserProfileRepo struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewUserProfileRepo(db *sqlx.DB, logger *slog.Logger) *UserProfileRepo {
	return &UserProfileRepo{db: db, logger: logger}
}

func (r *UserProfileRepo) Create(ctx context.Context, profile *domain.UserProfile) error {
	query := `
		INSERT INTO user_profiles (id, first_name, last_name, patronymic, avatar_url, phone, birthdate, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		profile.ID,
		profile.FirstName,
		profile.LastName,
		profile.Patronymic, profile.AvatarURL,
		profile.Phone,
		profile.Birthdate,
		profile.UpdatedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to create user profile",
			slog.String("user_id", profile.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create user profile: %w", err)
	}
	return nil
}

func (r *UserProfileRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.UserProfile, error) {
	query := `
		SELECT id, first_name, last_name, patronymic, avatar_url, phone, birthdate, updated_at
		FROM user_profiles
		WHERE id = $1
	`
	var profile domain.UserProfile
	err := r.db.GetContext(ctx, &profile, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user profile: %w", err)
	}
	return &profile, nil
}

func (r *UserProfileRepo) Update(ctx context.Context, profile *domain.UserProfile) error {
	query := `
		UPDATE user_profiles
		SET first_name = $2, last_name = $3, patronymic = $4, avatar_url = $5, phone = $6, birthdate = $7, updated_at = $8
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		profile.ID,
		profile.FirstName,
		profile.LastName,
		profile.Patronymic,
		profile.AvatarURL,
		profile.Phone,
		profile.Birthdate,
		profile.UpdatedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to update user profile",
			slog.String("user_id", profile.ID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "failed to get rows affected", slog.String("error", err.Error()))
	}
	if rows == 0 {
		return fmt.Errorf("user profile not found for update")
	}

	return nil
}

// --- MasterProfileRepo ---

type MasterProfileRepo struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewMasterProfileRepo(db *sqlx.DB, logger *slog.Logger) *MasterProfileRepo {
	return &MasterProfileRepo{db: db, logger: logger}
}

func (r *MasterProfileRepo) Create(ctx context.Context, mp *domain.MasterProfile) error {
	query := `
		INSERT INTO master_profiles (user_id, is_active, description, experience_years, rating, completed_orders, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		mp.UserID,
		mp.IsActive,
		mp.Description,
		mp.ExperienceYears,
		mp.Rating,
		mp.CompletedOrders,
		mp.UpdatedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to create master profile",
			slog.String("user_id", mp.UserID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create master profile: %w", err)
	}
	return nil
}

func (r *MasterProfileRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.MasterProfile, error) {
	query := `
		SELECT user_id, is_active, description, experience_years, rating, completed_orders, updated_at
		FROM master_profiles
		WHERE user_id = $1
	`
	var mp domain.MasterProfile
	err := r.db.GetContext(ctx, &mp, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find master profile: %w", err)
	}
	return &mp, nil
}

func (r *MasterProfileRepo) Update(ctx context.Context, mp *domain.MasterProfile) error {
	query := `
		UPDATE master_profiles
		SET is_active = $2, description = $3, experience_years = $4, rating = $5, completed_orders = $6, updated_at = $7
		WHERE user_id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		mp.UserID,
		mp.IsActive,
		mp.Description,
		mp.ExperienceYears,
		mp.Rating,
		mp.CompletedOrders,
		mp.UpdatedAt,
	)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to update master profile",
			slog.String("user_id", mp.UserID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to update master profile: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "failed to get rows affected", slog.String("error", err.Error()))
	}
	if rows == 0 {
		return fmt.Errorf("master profile not found for update")
	}

	return nil
}

func (r *MasterProfileRepo) FindActive(ctx context.Context) ([]*domain.MasterProfile, error) {
	query := `
		SELECT user_id, is_active, description, experience_years, rating, completed_orders, updated_at
		FROM master_profiles
		WHERE is_active = TRUE
		ORDER BY rating DESC
	`
	var profiles []*domain.MasterProfile
	err := r.db.SelectContext(ctx, &profiles, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find active masters: %w", err)
	}
	return profiles, nil
}

func (r *MasterProfileRepo) UpdateRating(ctx context.Context, userID uuid.UUID, rating float64) error {
	query := `
		UPDATE master_profiles
		SET rating = $2, updated_at = NOW()
		WHERE user_id = $1
	`
	result, err := r.db.ExecContext(ctx, query, userID, rating)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to update master rating",
			slog.String("user_id", userID.String()),
			slog.Float64("rating", rating),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to update master rating: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "failed to get rows affected", slog.String("error", err.Error()))
	}
	if rows == 0 {
		return fmt.Errorf("master profile not found for rating update")
	}

	return nil
}

func (r *MasterProfileRepo) IncrementCompletedOrders(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE master_profiles
		SET completed_orders = completed_orders + 1, updated_at = NOW()
		WHERE user_id = $1
	`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to increment completed orders",
			slog.String("user_id", userID.String()),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to increment completed orders: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "failed to get rows affected", slog.String("error", err.Error()))
	}
	if rows == 0 {
		return fmt.Errorf("master profile not found for order increment")
	}

	return nil
}

// --- UserRoleRepo ---

type UserRoleRepo struct {
	db     *sqlx.DB
	logger *slog.Logger
}

func NewUserRoleRepo(db *sqlx.DB, logger *slog.Logger) *UserRoleRepo {
	return &UserRoleRepo{db: db, logger: logger}
}

func (r *UserRoleRepo) GetRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT role FROM user_roles WHERE user_id = $1 ORDER BY role
	`
	var roles []string
	err := r.db.SelectContext(ctx, &roles, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	if roles == nil {
		roles = []string{}
	}
	return roles, nil
}

func (r *UserRoleRepo) AddRole(ctx context.Context, userID uuid.UUID, role string) error {
	query := `
		INSERT INTO user_roles (user_id, role)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, userID, role)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to add role",
			slog.String("user_id", userID.String()),
			slog.String("role", role),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to add role: %w", err)
	}
	return nil
}

func (r *UserRoleRepo) RemoveRole(ctx context.Context, userID uuid.UUID, role string) error {
	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role = $2
	`
	result, err := r.db.ExecContext(ctx, query, userID, role)
	if err != nil {
		r.logger.ErrorContext(ctx, "failed to remove role",
			slog.String("user_id", userID.String()),
			slog.String("role", role),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to remove role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		r.logger.WarnContext(ctx, "failed to get rows affected", slog.String("error", err.Error()))
	}
	if rows == 0 {
		r.logger.WarnContext(ctx, "role not found for removal",
			slog.String("user_id", userID.String()),
			slog.String("role", role),
		)
	}

	return nil
}
