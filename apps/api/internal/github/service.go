package github

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	oauth *OAuthService
	repo  *RepositoryService
}

func NewService(oauth *OAuthService, repo *RepositoryService) *Service {
	return &Service{oauth: oauth, repo: repo}
}

func (s *Service) StartConnect(ctx context.Context, input OAuthStartInput) (string, error) {
	return s.oauth.StartConnect(ctx, input)
}

func (s *Service) HandleCallback(ctx context.Context, input OAuthCallbackInput) (GitHubConnectionSummary, error) {
	return s.oauth.HandleCallback(ctx, input)
}

func (s *Service) ListGitHubRepositories(ctx context.Context, userID uuid.UUID) ([]GitHubRepository, error) {
	return s.repo.ListGitHubRepositories(ctx, userID)
}

func (s *Service) ImportRepositories(ctx context.Context, input ImportRepositoriesInput) ([]ImportedRepository, error) {
	return s.repo.ImportRepositories(ctx, input)
}
