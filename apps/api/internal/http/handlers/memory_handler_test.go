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
	servicememory "github.com/maab88/repomemory/apps/api/internal/services/memory"
)

type fakeMemoryQueryService struct {
	list    []servicememory.MemoryEntry
	detail  servicememory.MemoryEntry
	listErr error
	getErr  error
}

func (f *fakeMemoryQueryService) ListRepositoryMemory(_ context.Context, _, _ uuid.UUID) ([]servicememory.MemoryEntry, error) {
	return f.list, f.listErr
}

func (f *fakeMemoryQueryService) GetRepositoryMemoryEntry(_ context.Context, _, _, _ uuid.UUID) (servicememory.MemoryEntry, error) {
	return f.detail, f.getErr
}

func TestListRepositoryMemoryAuthorized(t *testing.T) {
	repoID := uuid.New()
	memID := uuid.New()
	memSvc := &fakeMemoryQueryService{
		list: []servicememory.MemoryEntry{
			{
				ID:             memID,
				RepositoryID:   repoID,
				OrganizationID: uuid.New(),
				Type:           "pr_summary",
				Title:          "PR summary",
				Summary:        "Summary",
				GeneratedBy:    "deterministic",
				CreatedAt:      time.Now().UTC(),
			},
		},
	}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, &noopRepositoryService{}, memSvc)

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/repositories/{repoId}/memory", h.ListRepositoryMemory)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/"+repoID.String()+"/memory", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestListRepositoryMemoryUnauthorizedForbidden(t *testing.T) {
	repoID := uuid.New()
	memSvc := &fakeMemoryQueryService{listErr: servicememory.ErrRepositoryForbidden}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, &noopRepositoryService{}, memSvc)

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/repositories/{repoId}/memory", h.ListRepositoryMemory)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/"+repoID.String()+"/memory", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestGetRepositoryMemoryDetailRespectsAuthz(t *testing.T) {
	repoID := uuid.New()
	memID := uuid.New()
	memSvc := &fakeMemoryQueryService{getErr: servicememory.ErrRepositoryForbidden}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, &noopRepositoryService{}, memSvc)

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/repositories/{repoId}/memory/{memoryId}", h.GetRepositoryMemoryDetail)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/repositories/"+repoID.String()+"/memory/"+memID.String(), nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}
