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

func TestListRepositoryDigestsSuccess(t *testing.T) {
	repoID := uuid.New()
	repoSvc := &fakeRepositoryQueryService{
		digests: []servicerepositories.Digest{
			{
				ID:           uuid.New(),
				RepositoryID: repoID,
				PeriodStart:  time.Now().UTC().AddDate(0, 0, -7),
				PeriodEnd:    time.Now().UTC(),
				Title:        "Weekly Digest",
				Summary:      "Summary",
				BodyMarkdown: "## Highlights",
				GeneratedBy:  "deterministic",
				CreatedAt:    time.Now().UTC(),
			},
		},
	}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{}, &noopSearchService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/repositories/{repoId}/digests", h.ListRepositoryDigests)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/"+repoID.String()+"/digests", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestListRepositoryDigestsForbidden(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{digestErr: servicerepositories.ErrRepositoryForbidden}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{}, &noopSearchService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/repositories/{repoId}/digests", h.ListRepositoryDigests)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/"+uuid.New().String()+"/digests", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGenerateRepositoryDigestQueued(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{
		job: servicejobs.Job{
			ID:        uuid.New(),
			JobType:   "repo.generate_digest",
			Status:    "queued",
			QueueName: "default",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{}, &noopSearchService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Post("/repositories/{repoId}/digests/generate", h.GenerateRepositoryDigest)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/repositories/"+uuid.New().String()+"/digests/generate", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}
