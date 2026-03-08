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
		r.Get("/organizations/{orgId}", deps.V1Handler.GetOrganization)
		r.Post("/github/connect/start", deps.V1Handler.StartGitHubConnect)
		r.Get("/github/callback", deps.V1Handler.GitHubCallback)
		r.Get("/github/repositories", deps.V1Handler.ListGitHubRepositories)
		r.Post("/github/repositories/import", deps.V1Handler.ImportGitHubRepositories)
		r.Get("/jobs/{jobId}", deps.V1Handler.GetJob)
		r.Get("/organizations/{orgId}/repositories", deps.V1Handler.ListOrganizationRepositories)
		r.Get("/repositories", deps.V1Handler.ListRepositories)
		r.Get("/repositories/{repoId}", deps.V1Handler.GetRepository)
		r.Post("/repositories/{repoId}/sync", deps.V1Handler.TriggerRepositorySync)
	})

	return r
}
