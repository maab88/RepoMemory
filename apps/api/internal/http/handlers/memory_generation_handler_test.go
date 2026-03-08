package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	servicejobs "github.com/maab88/repomemory/apps/api/internal/services/jobs"
	servicerepositories "github.com/maab88/repomemory/apps/api/internal/services/repositories"
)

func TestGenerateRepositoryMemoryReturnsQueuedJob(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{
		job: servicejobs.Job{
			ID:        uuid.New(),
			JobType:   "repo.generate_memory",
			Status:    "queued",
			QueueName: "default",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Post("/repositories/{repoId}/memory/generate", h.GenerateRepositoryMemory)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/repositories/"+uuid.New().String()+"/memory/generate", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}

func TestGenerateRepositoryMemoryForbidden(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{memoryErr: servicerepositories.ErrRepositoryForbidden}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Post("/repositories/{repoId}/memory/generate", h.GenerateRepositoryMemory)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/repositories/"+uuid.New().String()+"/memory/generate", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGenerateRepositoryMemoryNotFound(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{memoryErr: servicerepositories.ErrRepositoryNotFound}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Post("/repositories/{repoId}/memory/generate", h.GenerateRepositoryMemory)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/repositories/"+uuid.New().String()+"/memory/generate", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
