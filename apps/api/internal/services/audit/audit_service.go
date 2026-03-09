package audit

import (
	"context"

	"github.com/google/uuid"
	apimiddleware "github.com/maab88/repomemory/apps/api/internal/middleware"
	"github.com/maab88/repomemory/apps/api/internal/repositories"
)

type Repository interface {
	Create(ctx context.Context, input repositories.CreateAuditLogInput) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LogGitHubConnectionSuccess(ctx context.Context, userID uuid.UUID, organizationID *uuid.UUID, accountID uuid.UUID, githubLogin string) error {
	return s.repo.Create(ctx, repositories.CreateAuditLogInput{
		OrganizationID: organizationID,
		ActorUserID:    &userID,
		Action:         "github.connection_succeeded",
		EntityType:     "github_account",
		EntityID:       &accountID,
		Metadata: map[string]any{
			"githubLogin": githubLogin,
			"requestId":   apimiddleware.RequestIDFromContext(ctx),
		},
	})
}

func (s *Service) LogRepositorySyncTriggered(ctx context.Context, userID, organizationID, repositoryID, jobID uuid.UUID) error {
	return s.logJobTriggered(ctx, "repository.sync_triggered", userID, organizationID, repositoryID, jobID, "repo.initial_sync")
}

func (s *Service) LogMemoryGenerationTriggered(ctx context.Context, userID, organizationID, repositoryID, jobID uuid.UUID) error {
	return s.logJobTriggered(ctx, "repository.memory_generation_triggered", userID, organizationID, repositoryID, jobID, "repo.generate_memory")
}

func (s *Service) LogDigestGenerationTriggered(ctx context.Context, userID, organizationID, repositoryID, jobID uuid.UUID) error {
	return s.logJobTriggered(ctx, "repository.digest_generation_triggered", userID, organizationID, repositoryID, jobID, "repo.generate_digest")
}

func (s *Service) logJobTriggered(ctx context.Context, action string, userID, organizationID, repositoryID, jobID uuid.UUID, jobType string) error {
	return s.repo.Create(ctx, repositories.CreateAuditLogInput{
		OrganizationID: &organizationID,
		ActorUserID:    &userID,
		Action:         action,
		EntityType:     "repository",
		EntityID:       &repositoryID,
		Metadata: map[string]any{
			"jobId":     jobID.String(),
			"jobType":   jobType,
			"requestId": apimiddleware.RequestIDFromContext(ctx),
		},
	})
}
