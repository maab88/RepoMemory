package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/maab88/repomemory/apps/api/internal/http/handlers"
)

type Dependencies struct {
	AuthMiddleware func(next http.Handler) http.Handler
	V1Handler      *handlers.V1Handler
}

func New(deps Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", handlers.Health)

	r.Route("/v1", func(r chi.Router) {
		r.Use(deps.AuthMiddleware)
		r.Get("/me", deps.V1Handler.GetMe)
		r.Get("/organizations", deps.V1Handler.ListOrganizations)
		r.Post("/organizations", deps.V1Handler.CreateOrganization)
		r.Get("/organizations/{orgID}", deps.V1Handler.GetOrganization)
	})

	return r
}
