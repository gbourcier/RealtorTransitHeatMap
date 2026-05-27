package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/user"
)

const (
	minPasswordLength = 8

	loginMaxAttempts = 10
	loginWindow      = 15 * time.Minute
	loginLockout     = 15 * time.Minute
)

type Handlers struct {
	svc     *Service
	cfg     config.AuthConfig
	limiter *rateLimiter
}

func NewHandlers(svc *Service, cfg config.AuthConfig) *Handlers {
	return &Handlers{
		svc:     svc,
		cfg:     cfg,
		limiter: newRateLimiter(loginMaxAttempts, loginWindow, loginLockout),
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	IsActive  bool   `json:"isActive"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt.Unix(),
		UpdatedAt: u.UpdatedAt.Unix(),
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json", "err", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeTooManyRequests(w http.ResponseWriter, retryAfter time.Duration) {
	secs := int(retryAfter.Seconds())
	if secs < 1 {
		secs = 1
	}
	w.Header().Set("Retry-After", strconv.Itoa(secs))
	writeError(w, http.StatusTooManyRequests, "too many attempts, try again later")
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	ipKey := "ip:" + clientIP(r)
	userKey := "user:" + limitKey(req.Username)
	if d := h.limiter.retryAfter(ipKey); d > 0 {
		writeTooManyRequests(w, d)
		return
	}
	if d := h.limiter.retryAfter(userKey); d > 0 {
		writeTooManyRequests(w, d)
		return
	}

	raw, sess, u, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if errors.Is(err, ErrBadCredentials) || errors.Is(err, ErrInactive) {
		h.limiter.fail(ipKey)
		h.limiter.fail(userKey)
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		slog.Error("login failed", "err", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	h.limiter.reset(ipKey)
	h.limiter.reset(userKey)

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    raw,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		MaxAge:   int(time.Until(sess.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieName)
	if err == nil {
		_ = h.svc.Logout(r.Context(), cookie.Value)
	}
	http.SetCookie(w, expiredCookie(h.cfg.CookieSecure))
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	u := userFromContext(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}
