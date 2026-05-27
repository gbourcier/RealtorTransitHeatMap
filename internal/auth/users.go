package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/user"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandlers struct {
	users *user.Repository
	svc   *Service
}

func NewUserHandlers(users *user.Repository, svc *Service) *UserHandlers {
	return &UserHandlers{users: users, svc: svc}
}

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type updateUserRequest struct {
	Role     *string `json:"role,omitempty"`
	IsActive *bool   `json:"isActive,omitempty"`
	Password *string `json:"password,omitempty"`
}

func (h *UserHandlers) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		slog.Error("users list failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to list users")
		return
	}
	out := make([]userResponse, len(users))
	for i := range users {
		out[i] = toUserResponse(&users[i])
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *UserHandlers) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "username is required")
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}
	role := req.Role
	if role == "" {
		role = user.RoleUser
	}
	if role != user.RoleAdmin && role != user.RoleUser {
		writeError(w, http.StatusBadRequest, "role must be admin or user")
		return
	}

	hash, err := h.svc.HashPassword(req.Password)
	if err != nil {
		slog.Error("hash password failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	u := &user.User{
		Username:     req.Username,
		PasswordHash: hash,
		Role:         role,
		IsActive:     true,
	}
	if err := h.users.Create(r.Context(), u); err != nil {
		writeError(w, http.StatusConflict, "username already exists")
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(u))
}

func (h *UserHandlers) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if req.Role != nil && *req.Role != user.RoleAdmin && *req.Role != user.RoleUser {
		writeError(w, http.StatusBadRequest, "role must be admin or user")
		return
	}

	existing, err := h.users.GetByID(r.Context(), id)
	if errors.Is(err, user.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		slog.Error("users get failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	caller := userFromContext(r)
	if caller != nil && caller.ID == id {
		if req.IsActive != nil && !*req.IsActive {
			writeError(w, http.StatusBadRequest, "cannot deactivate your own account")
			return
		}
	}

	if existing.Role == user.RoleAdmin {
		needsCheck := (req.Role != nil && *req.Role != user.RoleAdmin) ||
			(req.IsActive != nil && !*req.IsActive)
		if needsCheck {
			count, countErr := h.users.CountAdmins(r.Context())
			if countErr != nil {
				slog.Error("count admins failed", "err", countErr)
				writeError(w, http.StatusInternalServerError, "failed to update user")
				return
			}
			if count <= 1 {
				writeError(w, http.StatusBadRequest, "cannot remove the last active admin")
				return
			}
		}
	}

	patch := user.Patch{
		Role:     req.Role,
		IsActive: req.IsActive,
	}
	if req.Password != nil {
		if *req.Password == "" {
			writeError(w, http.StatusBadRequest, "password cannot be empty")
			return
		}
		hash, hashErr := h.svc.HashPassword(*req.Password)
		if hashErr != nil {
			slog.Error("hash password failed", "err", hashErr)
			writeError(w, http.StatusInternalServerError, "failed to update user")
			return
		}
		patch.PasswordHash = &hash
	}

	updated, err := h.users.Update(r.Context(), id, patch)
	if errors.Is(err, user.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		slog.Error("users update failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(updated))
}

func (h *UserHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	caller := userFromContext(r)
	if caller != nil && caller.ID == id {
		writeError(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}

	existing, err := h.users.GetByID(r.Context(), id)
	if errors.Is(err, user.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		slog.Error("users get failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	if existing.Role == user.RoleAdmin {
		count, countErr := h.users.CountAdmins(r.Context())
		if countErr != nil {
			slog.Error("count admins failed", "err", countErr)
			writeError(w, http.StatusInternalServerError, "failed to delete user")
			return
		}
		if count <= 1 {
			writeError(w, http.StatusBadRequest, "cannot delete the last active admin")
			return
		}
	}

	if err := h.users.Delete(r.Context(), id); errors.Is(err, user.ErrNotFound) {
		writeError(w, http.StatusNotFound, "user not found")
		return
	} else if err != nil {
		slog.Error("users delete failed", "err", err)
		writeError(w, http.StatusInternalServerError, "failed to delete user")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
