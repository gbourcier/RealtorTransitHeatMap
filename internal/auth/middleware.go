package auth

import (
	"context"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/user"
)

const cookieName = "session"

type Guard struct {
	RequireAuth  func(http.Handler) http.Handler
	RequireAdmin func(http.Handler) http.Handler
}

func NewGuard(svc *Service) *Guard {
	g := &Guard{}

	g.RequireAuth = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			u, err := svc.Validate(r.Context(), cookie.Value)
			if err != nil {
				http.SetCookie(w, expiredCookie(svc.cfg.CookieSecure))
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			go svc.MarkSeen(context.WithoutCancel(r.Context()), u.ID)
			next.ServeHTTP(w, withUser(r, u))
		})
	}

	g.RequireAdmin = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := userFromContext(r)
			if u == nil || u.Role != user.RoleAdmin {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	return g
}

func expiredCookie(secure bool) *http.Cookie {
	return &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}
