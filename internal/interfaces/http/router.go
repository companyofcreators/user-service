package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/companyofcreators/user-service/pkg/header_auth"
)

// NewRouter creates and configures the chi router with all routes.
func NewRouter(handler *UserHandler, signer *header_auth.HeaderSigner) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(bodySizeLimiter(500 << 10)) // 500KB
	r.Use(chimiddleware.Heartbeat("/ping"))
	r.Use(signer.VerifyMiddleware)

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

// bodySizeLimiter returns middleware that wraps http.MaxBytesReader to limit
// request body size and prevent memory exhaustion attacks.
func bodySizeLimiter(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
