package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	appuser "github.com/companyofcreators/user-service/internal/application/user"
	domain "github.com/companyofcreators/user-service/internal/domain/user"
	"github.com/companyofcreators/user-service/pkg"
)

// UserHandler handles HTTP requests for user-related operations.
type UserHandler struct {
	getProfile         *appuser.GetProfileUseCase
	updateProfile      *appuser.UpdateProfileUseCase
	getMasterProfile   *appuser.GetMasterProfileUseCase
	updateMasterProfile *appuser.UpdateMasterProfileUseCase
	switchRole         *appuser.SwitchRoleUseCase
	roleRepo           domain.UserRoleRepository
	validator          *validator.Validate
	logger             *slog.Logger
}

func NewUserHandler(
	getProfile *appuser.GetProfileUseCase,
	updateProfile *appuser.UpdateProfileUseCase,
	getMasterProfile *appuser.GetMasterProfileUseCase,
	updateMasterProfile *appuser.UpdateMasterProfileUseCase,
	switchRole *appuser.SwitchRoleUseCase,
	roleRepo domain.UserRoleRepository,
	logger *slog.Logger,
) *UserHandler {
	return &UserHandler{
		getProfile:         getProfile,
		updateProfile:      updateProfile,
		getMasterProfile:   getMasterProfile,
		updateMasterProfile: updateMasterProfile,
		switchRole:         switchRole,
		roleRepo:           roleRepo,
		validator:          validator.New(),
		logger:             logger,
	}
}

// GetProfile GET /internal/users/{id}
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromURL(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID пользователя", err.Error())
		return
	}

	// Authorization check: X-User-Id must match {id} or user must be admin
	if err := h.authorizeRequest(r, userID, "admin"); err != nil {
		h.writeError(w, http.StatusForbidden, "доступ запрещён", err.Error())
		return
	}

	profile, err := h.getProfile.Execute(r.Context(), userID)
	if err != nil {
		if errors.Is(err, errUserNotFound) || err.Error() == "пользователь не найден" {
			h.writeError(w, http.StatusNotFound, "пользователь не найден", err.Error())
			return
		}
		h.logger.ErrorContext(r.Context(), "GetProfile failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "внутренняя ошибка", "не удалось получить профиль")
		return
	}

	h.writeJSON(w, http.StatusOK, NewFullProfileResponse(profile))
}

// UpdateProfile PATCH /internal/users/{id}
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromURL(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID пользователя", err.Error())
		return
	}

	// Authorization check: X-User-Id must match {id} or user must be admin
	if err := h.authorizeRequest(r, userID, "admin"); err != nil {
		h.writeError(w, http.StatusForbidden, "доступ запрещён", err.Error())
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "некорректное тело запроса", err.Error())
		return
	}
	defer r.Body.Close()

	if verrs := pkg.ValidateStruct(req); verrs != nil {
		pkg.WriteValidationErrors(w, verrs)
		return
	}

	profile, err := h.updateProfile.Execute(r.Context(), userID, req.ToDomain())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "UpdateProfile failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "внутренняя ошибка", "не удалось обновить профиль")
		return
	}

	h.writeJSON(w, http.StatusOK, NewUserProfileResponse(profile))
}

// GetMasterProfile GET /internal/masters/{id}
func (h *UserHandler) GetMasterProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromURL(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID пользователя", err.Error())
		return
	}

	// Authorization check: no restriction on reading master profiles (public info)
	// but we verify the X-User-Id header is present and valid
	callerIDStr := r.Header.Get("X-User-Id")
	if callerIDStr == "" {
		h.writeError(w, http.StatusForbidden, "доступ запрещён", "отсутствует заголовок X-User-Id")
		return
	}
	if _, err := uuid.Parse(callerIDStr); err != nil {
		h.writeError(w, http.StatusForbidden, "доступ запрещён", "недействительный заголовок X-User-Id")
		return
	}

	masterProfile, err := h.getMasterProfile.Execute(r.Context(), userID)
	if err != nil {
		if err.Error() == "профиль мастера не найден" {
			h.writeError(w, http.StatusNotFound, "профиль мастера не найден", err.Error())
			return
		}
		h.logger.ErrorContext(r.Context(), "GetMasterProfile failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "внутренняя ошибка", "не удалось получить профиль мастера")
		return
	}

	h.writeJSON(w, http.StatusOK, NewMasterProfileResponse(masterProfile))
}

// UpdateMasterProfile PATCH /internal/masters/{id}
func (h *UserHandler) UpdateMasterProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromURL(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID пользователя", err.Error())
		return
	}

	// Authorization check
	if err := h.authorizeRequest(r, userID, "admin"); err != nil {
		h.writeError(w, http.StatusForbidden, "доступ запрещён", err.Error())
		return
	}

	var req UpdateMasterProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "некорректное тело запроса", err.Error())
		return
	}
	defer r.Body.Close()

	if verrs := pkg.ValidateStruct(req); verrs != nil {
		pkg.WriteValidationErrors(w, verrs)
		return
	}

	masterProfile, err := h.updateMasterProfile.Execute(r.Context(), userID, req.ToDomain())
	if err != nil {
		h.logger.ErrorContext(r.Context(), "UpdateMasterProfile failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "внутренняя ошибка", "failed to update master profile")
		return
	}

	h.writeJSON(w, http.StatusOK, NewMasterProfileResponse(masterProfile))
}

// EnableMasterRole POST /internal/users/{id}/roles/master
func (h *UserHandler) EnableMasterRole(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromURL(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID пользователя", err.Error())
		return
	}

	// Authorization check
	if err := h.authorizeRequest(r, userID, "admin"); err != nil {
		h.writeError(w, http.StatusForbidden, "доступ запрещён", err.Error())
		return
	}

	masterProfile, err := h.switchRole.EnableMasterRole(r.Context(), userID)
	if err != nil {
		if err.Error() == "пользователь не найден" {
			h.writeError(w, http.StatusNotFound, "пользователь не найден", err.Error())
			return
		}
		h.logger.ErrorContext(r.Context(), "EnableMasterRole failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "внутренняя ошибка", "failed to enable master role")
		return
	}

	h.writeJSON(w, http.StatusOK, NewMasterProfileResponse(masterProfile))
}

// DisableMasterRole DELETE /internal/users/{id}/roles/master
func (h *UserHandler) DisableMasterRole(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromURL(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "недействительный ID пользователя", err.Error())
		return
	}

	// Authorization check
	if err := h.authorizeRequest(r, userID, "admin"); err != nil {
		h.writeError(w, http.StatusForbidden, "доступ запрещён", err.Error())
		return
	}

	if err := h.switchRole.DisableMasterRole(r.Context(), userID); err != nil {
		h.logger.ErrorContext(r.Context(), "DisableMasterRole failed", slog.String("error", err.Error()))
		h.writeError(w, http.StatusInternalServerError, "внутренняя ошибка", "failed to disable master role")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": "роль мастера отключена",
	})
}

// Health GET /internal/health
func (h *UserHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "user-service",
	})
}

// authorizeRequest checks if the X-User-Id header matches the target user ID
// or if the user has the specified bypass role (e.g., "admin").
func (h *UserHandler) authorizeRequest(r *http.Request, targetUserID uuid.UUID, bypassRole string) error {
	callerIDStr := r.Header.Get("X-User-Id")
	if callerIDStr == "" {
		return fmt.Errorf("отсутствует заголовок X-User-Id")
	}

	callerID, err := uuid.Parse(callerIDStr)
	if err != nil {
		return fmt.Errorf("invalid X-User-Id header: %w", err)
	}

	// Self-access is always allowed
	if callerID == targetUserID {
		return nil
	}

	// Check for bypass role
	roles, err := h.roleRepo.GetRoles(r.Context(), callerID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "failed to check caller roles",
			slog.String("caller_id", callerID.String()),
			slog.String("error", err.Error()),
		)
		// Deny access on role check failure (fail secure)
		return fmt.Errorf("unable to verify permissions")
	}

	for _, role := range roles {
		if role == bypassRole {
			return nil
		}
	}

	return fmt.Errorf("access denied")
}

// --- Helpers ---

var errUserNotFound = errors.New("пользователь не найден")

func getUserIDFromURL(r *http.Request) (uuid.UUID, error) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("missing id parameter")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid id format: %w", err)
	}

	return id, nil
}

func (h *UserHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *UserHandler) writeError(w http.ResponseWriter, status int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := ErrorResponse{
		Error:   message,
		Code:    status,
		Details: details,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Error("failed to encode error response", slog.String("error", err.Error()))
	}
}
