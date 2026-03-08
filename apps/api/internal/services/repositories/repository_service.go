package repositories

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/maab88/repomemory/apps/api/internal/db"
	repostore "github.com/maab88/repomemory/apps/api/internal/repositories"
	servicejobs "github.com/maab88/repomemory/apps/api/internal/services/jobs"
)

var (
	ErrRepositoryForbidden = errors.New("repository access denied")
	ErrRepositoryNotFound  = errors.New("repository not found")
)

type MembershipChecker interface {
	UserHasMembership(ctx context.Context, arg db.UserHasMembershipParams) (bool, error)
}

type JobEnqueuer interface {
	EnqueueRepositoryInitialSync(ctx context.Context, repositoryID, organizationID, triggeredByUserID uuid.UUID) (servicejobs.Job, error)
}

type Service struct {
	repoRepository      *repostore.RepositoryRepository
	syncStateRepository *repostore.RepositorySyncStateRepository
	membershipChecker   MembershipChecker
	jobEnqueuer         JobEnqueuer
}

func NewService(
	repoRepository *repostore.RepositoryRepository,
	syncStateRepository *repostore.RepositorySyncStateRepository,
	membershipChecker MembershipChecker,
	jobEnqueuer JobEnqueuer,
) *Service {
	return &Service{
		repoRepository:      repoRepository,
		syncStateRepository: syncStateRepository,
		membershipChecker:   membershipChecker,
		jobEnqueuer:         jobEnqueuer,
	}
}

type Repository struct {
	ID               uuid.UUID
	OrganizationID   uuid.UUID
	GitHubRepoID     string
	OwnerLogin       string
	Name             string
	FullName         string
	Private          bool
	DefaultBranch    string
	HTMLURL          string
	Description      string
	ImportedAt       time.Time
	LastSyncStatus   string
	LastSyncTime     *time.Time
	PullRequestCount int
	IssueCount       int
	MemoryEntryCount int
}

func (s *Service) ListOrganizationRepositories(ctx context.Context, userID, organizationID uuid.UUID) ([]Repository, error) {
	allowed, err := s.membershipChecker.UserHasMembership(ctx, db.UserHasMembershipParams{
		UserID:         userID,
		OrganizationID: organizationID,
	})
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrRepositoryForbidden
	}

	summaries, err := s.repoRepository.ListSummariesForOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	result := make([]Repository, 0, len(summaries))
	for _, item := range summaries {
		result = append(result, Repository{
			ID:               item.ID,
			OrganizationID:   item.OrganizationID,
			GitHubRepoID:     item.GitHubRepoID,
			OwnerLogin:       item.OwnerLogin,
			Name:             item.Name,
			FullName:         item.FullName,
			Private:          item.Private,
			DefaultBranch:    item.DefaultBranch,
			HTMLURL:          item.HTMLURL,
			Description:      item.Description,
			ImportedAt:       item.ImportedAt,
			LastSyncStatus:   item.LastSyncStatus,
			LastSyncTime:     item.LastSyncTime,
			PullRequestCount: item.PullRequestCount,
			IssueCount:       item.IssueCount,
			MemoryEntryCount: item.MemoryEntryCount,
		})
	}
	return result, nil
}

func (s *Service) ListRepositoriesForUser(ctx context.Context, userID uuid.UUID) ([]Repository, error) {
	summaries, err := s.repoRepository.ListSummariesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]Repository, 0, len(summaries))
	for _, item := range summaries {
		result = append(result, Repository{
			ID:               item.ID,
			OrganizationID:   item.OrganizationID,
			GitHubRepoID:     item.GitHubRepoID,
			OwnerLogin:       item.OwnerLogin,
			Name:             item.Name,
			FullName:         item.FullName,
			Private:          item.Private,
			DefaultBranch:    item.DefaultBranch,
			HTMLURL:          item.HTMLURL,
			Description:      item.Description,
			ImportedAt:       item.ImportedAt,
			LastSyncStatus:   item.LastSyncStatus,
			LastSyncTime:     item.LastSyncTime,
			PullRequestCount: item.PullRequestCount,
			IssueCount:       item.IssueCount,
			MemoryEntryCount: item.MemoryEntryCount,
		})
	}

	return result, nil
}

func (s *Service) GetRepository(ctx context.Context, userID, repositoryID uuid.UUID) (Repository, error) {
	repoRow, err := s.repoRepository.GetByID(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Repository{}, ErrRepositoryNotFound
		}
		return Repository{}, err
	}

	allowed, err := s.membershipChecker.UserHasMembership(ctx, db.UserHasMembershipParams{
		UserID:         userID,
		OrganizationID: repoRow.OrganizationID,
	})
	if err != nil {
		return Repository{}, err
	}
	if !allowed {
		return Repository{}, ErrRepositoryForbidden
	}

	summaries, err := s.repoRepository.ListSummariesForOrganization(ctx, repoRow.OrganizationID)
	if err != nil {
		return Repository{}, err
	}

	for _, item := range summaries {
		if item.ID == repositoryID {
			return Repository{
				ID:               item.ID,
				OrganizationID:   item.OrganizationID,
				GitHubRepoID:     item.GitHubRepoID,
				OwnerLogin:       item.OwnerLogin,
				Name:             item.Name,
				FullName:         item.FullName,
				Private:          item.Private,
				DefaultBranch:    item.DefaultBranch,
				HTMLURL:          item.HTMLURL,
				Description:      item.Description,
				ImportedAt:       item.ImportedAt,
				LastSyncStatus:   item.LastSyncStatus,
				LastSyncTime:     item.LastSyncTime,
				PullRequestCount: item.PullRequestCount,
				IssueCount:       item.IssueCount,
				MemoryEntryCount: item.MemoryEntryCount,
			}, nil
		}
	}

	lastSyncStatus := ""
	var lastSyncTime *time.Time
	syncState, err := s.syncStateRepository.GetByRepositoryID(ctx, repositoryID)
	if err == nil {
		lastSyncStatus = syncState.LastSyncStatus.String
		if syncState.LastSuccessfulSyncAt.Valid {
			t := syncState.LastSuccessfulSyncAt.Time.UTC()
			lastSyncTime = &t
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Repository{}, err
	}

	return Repository{
		ID:               repoRow.ID,
		OrganizationID:   repoRow.OrganizationID,
		GitHubRepoID:     strconv.FormatInt(repoRow.GithubRepoID, 10),
		OwnerLogin:       repoRow.OwnerLogin,
		Name:             repoRow.Name,
		FullName:         repoRow.FullName,
		Private:          repoRow.Private,
		DefaultBranch:    repoRow.DefaultBranch,
		HTMLURL:          repoRow.HtmlUrl,
		Description:      repoRow.Description.String,
		ImportedAt:       repoRow.ImportedAt.Time.UTC(),
		LastSyncStatus:   lastSyncStatus,
		LastSyncTime:     lastSyncTime,
		PullRequestCount: 0,
		IssueCount:       0,
		MemoryEntryCount: 0,
	}, nil
}

func (s *Service) TriggerInitialSync(ctx context.Context, userID, repositoryID uuid.UUID) (servicejobs.Job, error) {
	repo, err := s.GetRepository(ctx, userID, repositoryID)
	if err != nil {
		return servicejobs.Job{}, err
	}

	return s.jobEnqueuer.EnqueueRepositoryInitialSync(ctx, repo.ID, repo.OrganizationID, userID)
}
