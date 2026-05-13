package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func NewServer(addr string, scrapes ScrapeService, schedRepo ScheduleRepo, reloader ScheduleReloader, listings ListingService, transitSvc TransitService) *http.Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	h := &handlers{scrapes: scrapes}
	sh := &scheduleHandlers{repo: schedRepo, reloader: reloader}
	lh := &listingHandlers{svc: listings}
	th := &transitHandlers{svc: transitSvc}
	r.Route("/api", func(r chi.Router) {
		r.Route("/scrapes", func(r chi.Router) {
			r.Post("/", h.startScrape)
			r.Get("/", h.listScrapes)
			r.Get("/{id}", h.getScrape)
		})
		r.Route("/schedules", func(r chi.Router) {
			r.Get("/", sh.list)
			r.Post("/", sh.create)
			r.Get("/{id}", sh.get)
			r.Patch("/{id}", sh.update)
			r.Delete("/{id}", sh.delete)
		})
		r.Route("/listings", func(r chi.Router) {
			r.Get("/", lh.list)
			r.Get("/{board}/{mls}", lh.get)
		})
		r.Route("/transit", func(r chi.Router) {
			r.Post("/compute", th.compute)
			r.Post("/compute/{board}/{mls}", th.computeOne)
		})
	})

	return &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
}
