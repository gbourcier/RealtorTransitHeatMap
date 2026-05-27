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

func TestRouteGuard(t *testing.T) {
	sentinel := func(next http.Handler) http.Handler { return next }
	guard := &auth.Guard{
		RequireAuth:  sentinel,
		RequireAdmin: sentinel,
	}

	staticFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html></html>")},
	}

	r := NewRouter(staticFS, guard, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	sentinelPtr := reflect.ValueOf(sentinel).Pointer()

	var violations []string
	_ = chi.Walk(r, func(method, route string, _ http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/*" {
			return nil
		}
		key := method + " " + route
		if publicAllowlist[key] {
			return nil
		}
		for _, mw := range middlewares {
			if reflect.ValueOf(mw).Pointer() == sentinelPtr {
				return nil
			}
		}
		violations = append(violations, key)
		return nil
	})

	for _, v := range violations {
		t.Errorf("route %s has no auth middleware", v)
	}
}
