package api

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/gbourcier/RealtorTransitHeatMap/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(addr string, staticFS fs.FS, guard *auth.Guard, authH *auth.Handlers, userH *auth.UserHandlers, scrapes ScrapeService, gtfsRefresh GtfsRefreshService, schedRepo ScheduleRepo, reloader ScheduleReloader, dispatcher ScheduleDispatcher, listings ListingService, transitSvc TransitService, stopSvc StopService) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           NewRouter(staticFS, guard, authH, userH, scrapes, gtfsRefresh, schedRepo, reloader, dispatcher, listings, transitSvc, stopSvc),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func NewRouter(staticFS fs.FS, guard *auth.Guard, authH *auth.Handlers, userH *auth.UserHandlers, scrapes ScrapeService, gtfsRefresh GtfsRefreshService, schedRepo ScheduleRepo, reloader ScheduleReloader, dispatcher ScheduleDispatcher, listings ListingService, transitSvc TransitService, stopSvc StopService) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(60 * time.Second))

	h := &handlers{scrapes: scrapes}
	gh := &gtfsRefreshHandlers{svc: gtfsRefresh}
	sh := &scheduleHandlers{repo: schedRepo, reloader: reloader, dispatcher: dispatcher}
	lh := &listingHandlers{svc: listings}
	th := &transitHandlers{svc: transitSvc}
	sth := &stopHandlers{svc: stopSvc}

	rt := guard.Wrap(r)

	rt.PublicGet("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	rt.Route("/api", func(rt *auth.Routes) {
		rt.Route("/auth", func(rt *auth.Routes) {
			rt.PublicPost("/login", authH.Login)
			rt.Post("/logout", authH.Logout)
			rt.Get("/me", authH.Me)
		})

		rt.Route("/scrapes", func(rt *auth.Routes) {
			rt.AdminGet("/", h.listScrapes)
			rt.AdminGet("/{id}", h.getScrape)
		})

		rt.Route("/gtfs-refresh-runs", func(rt *auth.Routes) {
			rt.AdminPost("/", gh.start)
			rt.AdminGet("/", gh.list)
			rt.AdminGet("/{id}", gh.get)
		})

		rt.Route("/schedules", func(rt *auth.Routes) {
			rt.AdminGet("/", sh.list)
			rt.AdminPost("/", sh.create)
			rt.AdminGet("/{id}", sh.get)
			rt.AdminPatch("/{id}", sh.update)
			rt.AdminDelete("/{id}", sh.delete)
			rt.AdminPost("/{id}/run", sh.run)
		})

		rt.Route("/listings", func(rt *auth.Routes) {
			rt.Get("/", lh.list)
			rt.Get("/map", lh.mapList)
			rt.Get("/{board}/{mls}", lh.get)
		})

		rt.Route("/transit", func(rt *auth.Routes) {
			rt.AdminPost("/compute", th.compute)
			rt.AdminPost("/compute/{board}/{mls}", th.computeOne)
			rt.Get("/stops", sth.list)
		})

		rt.Route("/users", func(rt *auth.Routes) {
			rt.AdminGet("/", userH.List)
			rt.AdminPost("/", userH.Create)
			rt.AdminPatch("/{id}", userH.Update)
			rt.AdminDelete("/{id}", userH.Delete)
		})
	})

	rt.PublicHandle("/*", staticHandler(staticFS))

	return r
}
