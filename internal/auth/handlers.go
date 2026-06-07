package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/config"
	"github.com/gbourcier/RealtorTransitHeatMap/internal/user"
	"github.com/google/uuid"
)

type PreferenceLookup interface {
	DefaultFilterID(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
}

const (
	minPasswordLength = 8
	maxPasswordLength = 72

	maxAuthBodyBytes = 4096

	loginMaxAttempts = 10
	loginWindow      = 15 * time.Minute
	loginLockout     = 15 * time.Minute

	lastSeenThrottle = 15 * time.Minute
)

type Handlers struct {
	svc     *Service
	prefs   PreferenceLookup
	cfg     config.AuthConfig
	limiter *rateLimiter
}

func NewHandlers(svc *Service, prefs PreferenceLookup, cfg config.AuthConfig) *Handlers {
	return &Handlers{
		svc:     svc,
		prefs:   prefs,
		cfg:     cfg,
		limiter: newRateLimiter(loginMaxAttempts, loginWindow, loginLockout),
	}
}

type mePreferences struct {
	DefaultFilterID *string `json:"defaultFilterId"`
}

type meResponse struct {
	ID          string        `json:"id"`
	Username    string        `json:"username"`
	Role        string        `json:"role"`
	Preferences mePreferences `json:"preferences"`
}

func (h *Handlers) toMeResponse(ctx context.Context, u *user.User) meResponse {
	var defaultFilterID *string
	if id, err := h.prefs.DefaultFilterID(ctx, u.ID); err != nil {
		slog.Error("load default filter failed", "err", err)
	} else if id != nil {
		s := id.String()
		defaultFilterID = &s
	}
	return meResponse{
		ID:          u.ID.String(),
		Username:    u.Username,
		Role:        u.Role,
		Preferences: mePreferences{DefaultFilterID: defaultFilterID},
	}
}

func (h *Handlers) clientIP(r *http.Request) string {
	if h.cfg.TrustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			last := strings.TrimSpace(parts[len(parts)-1])
			if last != "" {
				return last
			}
		}
	}
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
	ID         string `json:"id"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	IsActive   bool   `json:"isActive"`
	LastSeenAt *int64 `json:"lastSeenAt"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

func toUserResponse(u *user.User) userResponse {
	var lastSeen *int64
	if u.LastSeenAt != nil {
		v := u.LastSeenAt.Unix()
		lastSeen = &v
	}
	return userResponse{
		ID:         u.ID.String(),
		Username:   u.Username,
		Role:       u.Role,
		IsActive:   u.IsActive,
		LastSeenAt: lastSeen,
		CreatedAt:  u.CreatedAt.Unix(),
		UpdatedAt:  u.UpdatedAt.Unix(),
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
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}
	if len(req.Password) > maxPasswordLength {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	ipKey := "ip:" + h.clientIP(r)
	if d := h.limiter.retryAfter(ipKey); d > 0 {
		writeTooManyRequests(w, d)
		return
	}

	raw, sess, u, err := h.svc.Login(r.Context(), req.Username, req.Password)
	if errors.Is(err, ErrBadCredentials) || errors.Is(err, ErrInactive) {
		h.limiter.fail(ipKey)
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		slog.Error("login failed", "err", err)
		writeError(w, http.StatusInternalServerError, "login failed")
		return
	}

	h.limiter.reset(ipKey)

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
	writeJSON(w, http.StatusOK, h.toMeResponse(r.Context(), u))
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
	writeJSON(w, http.StatusOK, h.toMeResponse(r.Context(), u))
}
