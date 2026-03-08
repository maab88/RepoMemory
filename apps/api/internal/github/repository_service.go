package github

import (
	"context"

	"github.com/google/uuid"
)

type RepositoryService struct {
	githubClient GitHubClient
	store        RepositoryStore
}

func NewRepositoryService(githubClient GitHubClient, store RepositoryStore) *RepositoryService {
	return &RepositoryService{githubClient: githubClient, store: store}
}

func (s *RepositoryService) ListGitHubRepositories(ctx context.Context, userID uuid.UUID) ([]GitHubRepository, error) {
	account, err := s.store.GetLatestGitHubAccountForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	repos, err := s.githubClient.ListRepositories(ctx, account.AccessTokenEncrypted)
	if err != nil {
		return nil, err
	}
	return repos, nil
}

func (s *RepositoryService) ImportRepositories(ctx context.Context, input ImportRepositoriesInput) ([]ImportedRepository, error) {
	if len(input.Repositories) == 0 {
		return nil, ErrImportRepositoriesEmpty
	}

	allowed, err := s.store.UserHasMembership(ctx, input.UserID, input.OrganizationID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrOrganizationAccessDenied
	}

	if _, err := s.store.GetLatestGitHubAccountForUser(ctx, input.UserID); err != nil {
		return nil, err
	}

	out := make([]ImportedRepository, 0, len(input.Repositories))
	for _, repo := range input.Repositories {
		if repo.GitHubRepoID == 0 || repo.OwnerLogin == "" || repo.Name == "" || repo.FullName == "" || repo.DefaultBranch == "" || repo.HTMLURL == "" {
			return nil, ErrInvalidRepositoryPayload
		}

		imported, err := s.store.UpsertRepositoryForOrganization(ctx, input.OrganizationID, repo)
		if err != nil {
			return nil, err
		}
		if err := s.store.EnsureRepositorySyncState(ctx, imported.ID); err != nil {
			return nil, err
		}
		if err := s.store.InsertRepositoryImportAuditLog(ctx, input.UserID, input.OrganizationID, imported.ID, repo); err != nil {
			return nil, err
		}
		out = append(out, imported)
	}

	return out, nil
}
