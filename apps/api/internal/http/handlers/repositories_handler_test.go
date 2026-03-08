package handlers

import (
	"context"
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

type fakeRepositoryQueryService struct {
	listForUser  []servicerepositories.Repository
	repositories []servicerepositories.Repository
	repository   servicerepositories.Repository
	job          servicejobs.Job
	listErr      error
	repoErr      error
	syncErr      error
}

func (f *fakeRepositoryQueryService) ListRepositoriesForUser(_ context.Context, _ uuid.UUID) ([]servicerepositories.Repository, error) {
	return f.listForUser, nil
}

func (f *fakeRepositoryQueryService) ListOrganizationRepositories(_ context.Context, _, _ uuid.UUID) ([]servicerepositories.Repository, error) {
	return f.repositories, f.listErr
}
func (f *fakeRepositoryQueryService) GetRepository(_ context.Context, _, _ uuid.UUID) (servicerepositories.Repository, error) {
	return f.repository, f.repoErr
}
func (f *fakeRepositoryQueryService) TriggerInitialSync(_ context.Context, _, _ uuid.UUID) (servicejobs.Job, error) {
	return f.job, f.syncErr
}

func TestListOrganizationRepositoriesReturnsPersistedRows(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{
		repositories: []servicerepositories.Repository{
			{
				ID:             uuid.New(),
				OrganizationID: uuid.New(),
				GitHubRepoID:   "123",
				FullName:       "octocat/repo-memory",
				DefaultBranch:  "main",
				HTMLURL:        "https://github.com/octocat/repo-memory",
				ImportedAt:     time.Now().UTC(),
			},
		},
	}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/organizations/{orgId}/repositories", h.ListOrganizationRepositories)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+uuid.New().String()+"/repositories", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestListRepositoriesReturnsPersistedValues(t *testing.T) {
	lastSync := time.Now().UTC()
	repoSvc := &fakeRepositoryQueryService{
		listForUser: []servicerepositories.Repository{
			{
				ID:               uuid.New(),
				OrganizationID:   uuid.New(),
				GitHubRepoID:     "123",
				OwnerLogin:       "octocat",
				Name:             "repo-memory",
				FullName:         "octocat/repo-memory",
				Private:          true,
				DefaultBranch:    "main",
				HTMLURL:          "https://github.com/octocat/repo-memory",
				ImportedAt:       time.Now().UTC(),
				LastSyncStatus:   "succeeded",
				LastSyncTime:     &lastSync,
				PullRequestCount: 12,
				IssueCount:       8,
			},
		},
	}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/repositories", h.ListRepositories)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/repositories", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestTriggerRepositorySyncReturnsJobID(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{
		job: servicejobs.Job{
			ID:        uuid.New(),
			JobType:   "repo.initial_sync",
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
		r.Post("/repositories/{repoId}/sync", h.TriggerRepositorySync)
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/repositories/"+uuid.New().String()+"/sync", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}

func TestListOrganizationRepositoriesForbidden(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{listErr: servicerepositories.ErrRepositoryForbidden}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/organizations/{orgId}/repositories", h.ListOrganizationRepositories)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+uuid.New().String()+"/repositories", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGetRepositoryDetailNotFound(t *testing.T) {
	repoSvc := &fakeRepositoryQueryService{repoErr: servicerepositories.ErrRepositoryNotFound}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, repoSvc, &noopMemoryService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/repositories/{repoId}", h.GetRepository)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/"+uuid.New().String(), nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
