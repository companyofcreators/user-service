package http

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates and configures the chi router with all routes.
func NewRouter(handler *UserHandler) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Heartbeat("/ping"))

	// Health check
	r.Get("/internal/health", handler.Health)

	// User profile routes
	r.Route("/internal/users/{id}", func(r chi.Router) {
		r.Get("/", handler.GetProfile)
		r.Patch("/", handler.UpdateProfile)

		// Role switching
		r.Post("/roles/master", handler.EnableMasterRole)
		r.Delete("/roles/master", handler.DisableMasterRole)
	})

	// Master profile routes
	r.Route("/internal/masters/{id}", func(r chi.Router) {
		r.Get("/", handler.GetMasterProfile)
		r.Patch("/", handler.UpdateMasterProfile)
	})

	return r
}
