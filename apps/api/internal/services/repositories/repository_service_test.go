package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/db"
	repostore "github.com/maab88/repomemory/apps/api/internal/repositories"
	servicejobs "github.com/maab88/repomemory/apps/api/internal/services/jobs"
)

type fakeMembershipChecker struct {
	allowed bool
}

func (f *fakeMembershipChecker) UserHasMembership(context.Context, db.UserHasMembershipParams) (bool, error) {
	return f.allowed, nil
}

type fakeJobEnqueuer struct {
	lastRepoID uuid.UUID
}

func (f *fakeJobEnqueuer) EnqueueRepositoryInitialSync(_ context.Context, repositoryID, _ uuid.UUID, _ uuid.UUID) (servicejobs.Job, error) {
	f.lastRepoID = repositoryID
	return servicejobs.Job{ID: uuid.New(), Status: "queued"}, nil
}

func (f *fakeJobEnqueuer) EnqueueRepositoryGenerateMemory(_ context.Context, repositoryID, _ uuid.UUID, _ uuid.UUID) (servicejobs.Job, error) {
	f.lastRepoID = repositoryID
	return servicejobs.Job{ID: uuid.New(), Status: "queued"}, nil
}
func (f *fakeJobEnqueuer) EnqueueRepositoryGenerateDigest(_ context.Context, repositoryID, _ uuid.UUID, _ uuid.UUID) (servicejobs.Job, error) {
	f.lastRepoID = repositoryID
	return servicejobs.Job{ID: uuid.New(), Status: "queued"}, nil
}

func TestListOrganizationRepositoriesForbidden(t *testing.T) {
	svc := NewService(
		&repostore.RepositoryRepository{},
		&repostore.RepositorySyncStateRepository{},
		&repostore.DigestRepository{},
		&fakeMembershipChecker{allowed: false},
		&fakeJobEnqueuer{},
	)

	_, err := svc.ListOrganizationRepositories(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrRepositoryForbidden) {
		t.Fatalf("expected ErrRepositoryForbidden, got %v", err)
	}
}

func TestRepositoryMapping(t *testing.T) {
	now := time.Now().UTC()
	item := repostore.RepositorySummary{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		GitHubRepoID:   "123",
		OwnerLogin:     "octocat",
		Name:           "repo-memory",
		FullName:       "octocat/repo-memory",
		Private:        true,
		DefaultBranch:  "main",
		HTMLURL:        "https://github.com/octocat/repo-memory",
		ImportedAt:     now,
	}

	got := Repository{
		ID:             item.ID,
		OrganizationID: item.OrganizationID,
		GitHubRepoID:   item.GitHubRepoID,
		OwnerLogin:     item.OwnerLogin,
		Name:           item.Name,
		FullName:       item.FullName,
		Private:        item.Private,
		DefaultBranch:  item.DefaultBranch,
		HTMLURL:        item.HTMLURL,
		ImportedAt:     item.ImportedAt,
	}

	if got.FullName != "octocat/repo-memory" {
		t.Fatalf("unexpected full name: %s", got.FullName)
	}
}
