package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/maab88/repomemory/apps/api/internal/http/handlers"
)

func New() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", handlers.Health)
	return r
}
