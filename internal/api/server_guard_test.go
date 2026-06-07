package api

import (
	"net/http"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/auth"
	"github.com/go-chi/chi/v5"
)

var publicAllowlist = map[string]bool{
	"GET /healthz":         true,
	"POST /api/auth/login": true,
}

var authOnlyAllowlist = map[string]bool{
	"POST /api/auth/logout":               true,
	"GET /api/auth/me":                    true,
	"GET /api/listings/":                  true,
	"GET /api/listings/map":               true,
	"GET /api/listings/{board}/{mls}":     true,
	"GET /api/favorites/":                 true,
	"POST /api/favorites/":                true,
	"DELETE /api/favorites/":              true,
	"DELETE /api/favorites/{board}/{mls}": true,
	"GET /api/saved-filters/":             true,
	"POST /api/saved-filters/":            true,
	"PATCH /api/saved-filters/{id}":       true,
	"DELETE /api/saved-filters/{id}":      true,
	"PATCH /api/preferences/":             true,
	"GET /api/transit/stops":              true,
}

func TestRouteGuard(t *testing.T) {
	authSentinel := func(next http.Handler) http.Handler { return next }
	adminSentinel := func(next http.Handler) http.Handler { return next }
	guard := &auth.Guard{
		RequireAuth:  authSentinel,
		RequireAdmin: adminSentinel,
	}

	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}

	r := NewRouter(staticFS, guard, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	authPtr := reflect.ValueOf(authSentinel).Pointer()
	adminPtr := reflect.ValueOf(adminSentinel).Pointer()

	_ = chi.Walk(r, func(method, route string, _ http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/*" {
			return nil
		}
		key := method + " " + route

		var hasAuth, hasAdmin bool
		for _, mw := range middlewares {
			switch reflect.ValueOf(mw).Pointer() {
			case authPtr:
				hasAuth = true
			case adminPtr:
				hasAdmin = true
			}
		}

		if publicAllowlist[key] {
			if hasAuth || hasAdmin {
				t.Errorf("route %s is in the public allowlist but has auth middleware", key)
			}
			return nil
		}

		if !hasAuth {
			t.Errorf("route %s has no auth middleware", key)
		}
		if !authOnlyAllowlist[key] && !hasAdmin {
			t.Errorf("route %s is not admin-gated and not in the authOnly allowlist", key)
		}
		return nil
	})
}
