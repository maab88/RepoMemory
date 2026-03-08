package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	dbpkg "github.com/maab88/repomemory/apps/api/internal/db"
	jobdefs "github.com/maab88/repomemory/apps/api/internal/jobs"
)

var (
	ErrJobForbidden = errors.New("job access denied")
	ErrJobNotFound  = errors.New("job not found")
)

type JobRepository interface {
	GetByID(ctx context.Context, jobID uuid.UUID) (dbpkg.Job, error)
}

type MembershipChecker interface {
	UserHasMembership(ctx context.Context, arg dbpkg.UserHasMembershipParams) (bool, error)
}

type RepositoryReader interface {
	GetRepositoryByID(ctx context.Context, id uuid.UUID) (dbpkg.Repository, error)
}

type RepoInitialSyncEnqueuer interface {
	EnqueueRepoInitialSync(ctx context.Context, payload jobdefs.RepoInitialSyncPayload) (jobdefs.JobRecord, error)
	EnqueueRepoGenerateMemory(ctx context.Context, payload jobdefs.RepoGenerateMemoryPayload) (jobdefs.JobRecord, error)
	EnqueueRepoGenerateDigest(ctx context.Context, payload jobdefs.RepoGenerateDigestPayload) (jobdefs.JobRecord, error)
}

type Service struct {
	jobRepo     JobRepository
	membership  MembershipChecker
	repositoryQ RepositoryReader
	enqueuer    RepoInitialSyncEnqueuer
}

func NewService(jobRepo JobRepository, membership MembershipChecker, repositoryQ RepositoryReader, enqueuer RepoInitialSyncEnqueuer) *Service {
	return &Service{
		jobRepo:     jobRepo,
		membership:  membership,
		repositoryQ: repositoryQ,
		enqueuer:    enqueuer,
	}
}

type Job struct {
	ID        uuid.UUID
	JobType   string
	Status    string
	QueueName string
	Attempts  int32
	LastError *string
	Payload   json.RawMessage

	StartedAt  *time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (s *Service) GetJob(ctx context.Context, userID, jobID uuid.UUID) (Job, error) {
	row, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Job{}, ErrJobNotFound
		}
		return Job{}, err
	}

	allowed, err := s.isUserAuthorizedForJob(ctx, userID, row)
	if err != nil {
		return Job{}, err
	}
	if !allowed {
		return Job{}, ErrJobForbidden
	}

	return mapJob(row), nil
}

func (s *Service) EnqueueRepositoryInitialSync(
	ctx context.Context,
	repositoryID, organizationID, triggeredByUserID uuid.UUID,
) (Job, error) {
	created, err := s.enqueuer.EnqueueRepoInitialSync(ctx, jobdefs.RepoInitialSyncPayload{
		RepositoryID:      repositoryID,
		OrganizationID:    organizationID,
		TriggeredByUserID: triggeredByUserID,
	})
	if err != nil {
		return Job{}, err
	}

	now := time.Now().UTC()
	return Job{
		ID:        created.ID,
		JobType:   created.JobType,
		Status:    created.Status,
		QueueName: created.QueueName,
		Attempts:  created.Attempts,
		LastError: created.LastError,
		Payload:   created.Payload,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Service) EnqueueRepositoryGenerateMemory(
	ctx context.Context,
	repositoryID, organizationID, triggeredByUserID uuid.UUID,
) (Job, error) {
	created, err := s.enqueuer.EnqueueRepoGenerateMemory(ctx, jobdefs.RepoGenerateMemoryPayload{
		RepositoryID:      repositoryID,
		OrganizationID:    organizationID,
		TriggeredByUserID: triggeredByUserID,
	})
	if err != nil {
		return Job{}, err
	}

	now := time.Now().UTC()
	return Job{
		ID:        created.ID,
		JobType:   created.JobType,
		Status:    created.Status,
		QueueName: created.QueueName,
		Attempts:  created.Attempts,
		LastError: created.LastError,
		Payload:   created.Payload,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Service) EnqueueRepositoryGenerateDigest(
	ctx context.Context,
	repositoryID, organizationID, triggeredByUserID uuid.UUID,
) (Job, error) {
	created, err := s.enqueuer.EnqueueRepoGenerateDigest(ctx, jobdefs.RepoGenerateDigestPayload{
		RepositoryID:      repositoryID,
		OrganizationID:    organizationID,
		TriggeredByUserID: triggeredByUserID,
	})
	if err != nil {
		return Job{}, err
	}

	now := time.Now().UTC()
	return Job{
		ID:        created.ID,
		JobType:   created.JobType,
		Status:    created.Status,
		QueueName: created.QueueName,
		Attempts:  created.Attempts,
		LastError: created.LastError,
		Payload:   created.Payload,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *Service) isUserAuthorizedForJob(ctx context.Context, userID uuid.UUID, row dbpkg.Job) (bool, error) {
	orgID := row.OrganizationID
	if orgID == nil && row.RepositoryID != nil {
		repo, err := s.repositoryQ.GetRepositoryByID(ctx, *row.RepositoryID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return false, nil
			}
			return false, err
		}
		orgID = &repo.OrganizationID
	}

	if orgID == nil {
		return false, nil
	}

	return s.membership.UserHasMembership(ctx, dbpkg.UserHasMembershipParams{
		UserID:         userID,
		OrganizationID: *orgID,
	})
}

func mapJob(row dbpkg.Job) Job {
	var lastError *string
	if row.LastError.Valid {
		lastError = &row.LastError.String
	}
	var startedAt *time.Time
	if row.StartedAt.Valid {
		t := row.StartedAt.Time.UTC()
		startedAt = &t
	}
	var finishedAt *time.Time
	if row.FinishedAt.Valid {
		t := row.FinishedAt.Time.UTC()
		finishedAt = &t
	}

	return Job{
		ID:         row.ID,
		JobType:    row.JobType,
		Status:     row.Status,
		QueueName:  row.QueueName.String,
		Attempts:   row.Attempts,
		LastError:  lastError,
		Payload:    row.Payload,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		CreatedAt:  row.CreatedAt.Time.UTC(),
		UpdatedAt:  row.UpdatedAt.Time.UTC(),
	}
}
