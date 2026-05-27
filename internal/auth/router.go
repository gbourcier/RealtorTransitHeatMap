package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Routes struct {
	mux  chi.Router
	auth func(http.Handler) http.Handler
	adm  func(http.Handler) http.Handler
}

func (g *Guard) Wrap(r chi.Router) *Routes {
	return &Routes{mux: r, auth: g.RequireAuth, adm: g.RequireAdmin}
}

func (rt *Routes) Route(pattern string, fn func(*Routes)) {
	rt.mux.Route(pattern, func(r chi.Router) {
		fn(&Routes{mux: r, auth: rt.auth, adm: rt.adm})
	})
}

func (rt *Routes) Get(pattern string, h http.HandlerFunc) {
	rt.mux.With(rt.auth).Get(pattern, h)
}

func (rt *Routes) Post(pattern string, h http.HandlerFunc) {
	rt.mux.With(rt.auth).Post(pattern, h)
}

func (rt *Routes) Patch(pattern string, h http.HandlerFunc) {
	rt.mux.With(rt.auth).Patch(pattern, h)
}

func (rt *Routes) Delete(pattern string, h http.HandlerFunc) {
	rt.mux.With(rt.auth).Delete(pattern, h)
}

func (rt *Routes) AdminGet(pattern string, h http.HandlerFunc) {
	rt.mux.With(rt.auth, rt.adm).Get(pattern, h)
}

func (rt *Routes) AdminPost(pattern string, h http.HandlerFunc) {
	rt.mux.With(rt.auth, rt.adm).Post(pattern, h)
}

func (rt *Routes) AdminPatch(pattern string, h http.HandlerFunc) {
	rt.mux.With(rt.auth, rt.adm).Patch(pattern, h)
}

func (rt *Routes) AdminDelete(pattern string, h http.HandlerFunc) {
	rt.mux.With(rt.auth, rt.adm).Delete(pattern, h)
}

func (rt *Routes) PublicGet(pattern string, h http.HandlerFunc) {
	rt.mux.Get(pattern, h)
}

func (rt *Routes) PublicPost(pattern string, h http.HandlerFunc) {
	rt.mux.Post(pattern, h)
}

func (rt *Routes) PublicHandle(pattern string, h http.Handler) {
	rt.mux.Handle(pattern, h)
}
