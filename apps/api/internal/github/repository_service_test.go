package github

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepoClient struct {
	repositories []GitHubRepository
	err          error
}

func (f *fakeRepoClient) ExchangeCode(_ context.Context, _, _ string) (GitHubToken, error) {
	return GitHubToken{}, nil
}

func (f *fakeRepoClient) GetViewer(_ context.Context, _ string) (GitHubUser, error) {
	return GitHubUser{}, nil
}

func (f *fakeRepoClient) ListRepositories(_ context.Context, _ string) ([]GitHubRepository, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.repositories, nil
}

type fakeRepoStore struct {
	allowMembership bool
	accountErr      error
	membershipErr   error
	upsertErr       error
	stateErr        error
	auditErr        error
	account         ConnectedGitHubAccount
	upsertCount     int
}

func (f *fakeRepoStore) UpsertGitHubAccount(_ context.Context, _ UpsertGitHubAccountInput) (GitHubConnectionSummary, error) {
	return GitHubConnectionSummary{}, nil
}

func (f *fakeRepoStore) UserHasMembership(_ context.Context, _, _ uuid.UUID) (bool, error) {
	if f.membershipErr != nil {
		return false, f.membershipErr
	}
	return f.allowMembership, nil
}

func (f *fakeRepoStore) GetLatestGitHubAccountForUser(_ context.Context, _ uuid.UUID) (ConnectedGitHubAccount, error) {
	if f.accountErr != nil {
		return ConnectedGitHubAccount{}, f.accountErr
	}
	return f.account, nil
}

func (f *fakeRepoStore) UpsertRepositoryForOrganization(_ context.Context, organizationID uuid.UUID, repo GitHubRepository) (ImportedRepository, error) {
	if f.upsertErr != nil {
		return ImportedRepository{}, f.upsertErr
	}
	f.upsertCount++
	return ImportedRepository{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		GitHubRepoID:   "123",
		OwnerLogin:     repo.OwnerLogin,
		Name:           repo.Name,
		FullName:       repo.FullName,
		Private:        repo.Private,
		DefaultBranch:  repo.DefaultBranch,
		HTMLURL:        repo.HTMLURL,
	}, nil
}

func (f *fakeRepoStore) EnsureRepositorySyncState(_ context.Context, _ uuid.UUID) error {
	return f.stateErr
}

func (f *fakeRepoStore) InsertRepositoryImportAuditLog(_ context.Context, _, _, _ uuid.UUID, _ GitHubRepository) error {
	return f.auditErr
}

func TestListGitHubRepositoriesWithMockClient(t *testing.T) {
	service := NewRepositoryService(&fakeRepoClient{repositories: []GitHubRepository{{GitHubRepoID: 1, OwnerLogin: "octocat", Name: "repo", FullName: "octocat/repo", DefaultBranch: "main", HTMLURL: "https://github.com/octocat/repo"}}}, &fakeRepoStore{account: ConnectedGitHubAccount{UserID: uuid.New(), AccessTokenEncrypted: "tok"}})

	repos, err := service.ListGitHubRepositories(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected one repo, got %d", len(repos))
	}
}

func TestImportOneAndMultipleRepositories(t *testing.T) {
	store := &fakeRepoStore{allowMembership: true, account: ConnectedGitHubAccount{AccessTokenEncrypted: "tok"}}
	service := NewRepositoryService(&fakeRepoClient{}, store)
	orgID := uuid.New()
	userID := uuid.New()

	one, err := service.ImportRepositories(context.Background(), ImportRepositoriesInput{UserID: userID, OrganizationID: orgID, Repositories: []GitHubRepository{{GitHubRepoID: 1, OwnerLogin: "octo", Name: "r1", FullName: "octo/r1", Private: true, DefaultBranch: "main", HTMLURL: "https://github.com/octo/r1"}}})
	if err != nil || len(one) != 1 {
		t.Fatalf("expected single import success, err=%v len=%d", err, len(one))
	}

	multi, err := service.ImportRepositories(context.Background(), ImportRepositoriesInput{UserID: userID, OrganizationID: orgID, Repositories: []GitHubRepository{{GitHubRepoID: 2, OwnerLogin: "octo", Name: "r2", FullName: "octo/r2", Private: true, DefaultBranch: "main", HTMLURL: "https://github.com/octo/r2"}, {GitHubRepoID: 3, OwnerLogin: "octo", Name: "r3", FullName: "octo/r3", Private: false, DefaultBranch: "main", HTMLURL: "https://github.com/octo/r3"}}})
	if err != nil || len(multi) != 2 {
		t.Fatalf("expected multi import success, err=%v len=%d", err, len(multi))
	}
}

func TestImportDuplicateDoesNotDuplicateRows(t *testing.T) {
	store := &fakeRepoStore{allowMembership: true, account: ConnectedGitHubAccount{AccessTokenEncrypted: "tok"}}
	service := NewRepositoryService(&fakeRepoClient{}, store)
	input := ImportRepositoriesInput{UserID: uuid.New(), OrganizationID: uuid.New(), Repositories: []GitHubRepository{{GitHubRepoID: 1, OwnerLogin: "octo", Name: "r1", FullName: "octo/r1", Private: true, DefaultBranch: "main", HTMLURL: "https://github.com/octo/r1"}}}

	if _, err := service.ImportRepositories(context.Background(), input); err != nil {
		t.Fatalf("first import failed: %v", err)
	}
	if _, err := service.ImportRepositories(context.Background(), input); err != nil {
		t.Fatalf("second import failed: %v", err)
	}
	if store.upsertCount != 2 {
		t.Fatalf("expected two upsert calls (safe re-import), got %d", store.upsertCount)
	}
}

func TestImportUnauthorizedOrganizationRejected(t *testing.T) {
	service := NewRepositoryService(&fakeRepoClient{}, &fakeRepoStore{allowMembership: false, account: ConnectedGitHubAccount{AccessTokenEncrypted: "tok"}})
	_, err := service.ImportRepositories(context.Background(), ImportRepositoriesInput{UserID: uuid.New(), OrganizationID: uuid.New(), Repositories: []GitHubRepository{{GitHubRepoID: 1, OwnerLogin: "octo", Name: "r1", FullName: "octo/r1", Private: true, DefaultBranch: "main", HTMLURL: "https://github.com/octo/r1"}}})
	if !errors.Is(err, ErrOrganizationAccessDenied) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestMissingGitHubConnectionRejected(t *testing.T) {
	service := NewRepositoryService(&fakeRepoClient{}, &fakeRepoStore{allowMembership: true, accountErr: ErrGitHubNotConnected})
	_, err := service.ImportRepositories(context.Background(), ImportRepositoriesInput{UserID: uuid.New(), OrganizationID: uuid.New(), Repositories: []GitHubRepository{{GitHubRepoID: 1, OwnerLogin: "octo", Name: "r1", FullName: "octo/r1", Private: true, DefaultBranch: "main", HTMLURL: "https://github.com/octo/r1"}}})
	if !errors.Is(err, ErrGitHubNotConnected) {
		t.Fatalf("expected not connected, got %v", err)
	}
}
