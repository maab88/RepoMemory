package audit

import (
	"context"
	"testing"

	"github.com/google/uuid"
	apimiddleware "github.com/maab88/repomemory/apps/api/internal/middleware"
	"github.com/maab88/repomemory/apps/api/internal/repositories"
)

type fakeAuditRepo struct {
	last repositories.CreateAuditLogInput
}

func (f *fakeAuditRepo) Create(_ context.Context, input repositories.CreateAuditLogInput) error {
	f.last = input
	return nil
}

func TestLogRepositorySyncTriggeredIncludesRequestID(t *testing.T) {
	repo := &fakeAuditRepo{}
	svc := NewService(repo)

	ctx := apimiddleware.ContextWithRequestID(context.Background(), "req-123")
	userID := uuid.New()
	orgID := uuid.New()
	repoID := uuid.New()
	jobID := uuid.New()

	if err := svc.LogRepositorySyncTriggered(ctx, userID, orgID, repoID, jobID); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.last.Action != "repository.sync_triggered" {
		t.Fatalf("unexpected action: %s", repo.last.Action)
	}
	if repo.last.Metadata["requestId"] != "req-123" {
		t.Fatalf("expected requestId metadata, got %+v", repo.last.Metadata)
	}
}

func TestLogGitHubConnectionSuccess(t *testing.T) {
	repo := &fakeAuditRepo{}
	svc := NewService(repo)

	orgID := uuid.New()
	accountID := uuid.New()
	if err := svc.LogGitHubConnectionSuccess(context.Background(), uuid.New(), &orgID, accountID, "octocat"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.last.Action != "github.connection_succeeded" {
		t.Fatalf("unexpected action: %s", repo.last.Action)
	}
	if repo.last.EntityID == nil || *repo.last.EntityID != accountID {
		t.Fatalf("unexpected entity id: %v", repo.last.EntityID)
	}
}
