package auth

import (
	"context"
	"net/http"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/user"
)

type ctxKey struct{}

func userFromContext(r *http.Request) *user.User {
	u, _ := r.Context().Value(ctxKey{}).(*user.User)
	return u
}

func UserFromContext(r *http.Request) *user.User {
	return userFromContext(r)
}

func withUser(r *http.Request, u *user.User) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), ctxKey{}, u))
}

func WithUser(r *http.Request, u *user.User) *http.Request {
	return withUser(r, u)
}
