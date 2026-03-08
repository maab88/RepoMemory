package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maab88/repomemory/apps/api/internal/db"
	"github.com/maab88/repomemory/apps/api/internal/jobs"
)

type JobRepository struct {
	queries *db.Queries
}

func NewJobRepository(queries *db.Queries) *JobRepository {
	return &JobRepository{queries: queries}
}

func (r *JobRepository) CreateJob(ctx context.Context, input jobs.CreateJobInput) (jobs.JobRecord, error) {
	row, err := r.queries.InsertJob(ctx, db.InsertJobParams{
		OrganizationID: input.OrganizationID,
		RepositoryID:   input.RepositoryID,
		JobType:        input.JobType,
		Status:         jobs.StatusQueued,
		QueueName:      pgtype.Text{String: input.QueueName, Valid: input.QueueName != ""},
		Attempts:       0,
		LastError:      pgtype.Text{},
		Payload:        input.Payload,
		StartedAt:      pgtype.Timestamptz{},
		FinishedAt:     pgtype.Timestamptz{},
	})
	if err != nil {
		return jobs.JobRecord{}, err
	}
	return toJobRecord(row), nil
}

func (r *JobRepository) UpdateJobLifecycle(ctx context.Context, input jobs.UpdateJobLifecycleInput) (jobs.JobRecord, error) {
	lastError := pgtype.Text{}
	if input.LastError != nil {
		lastError = pgtype.Text{String: *input.LastError, Valid: true}
	}

	row, err := r.queries.UpdateJobStatus(ctx, db.UpdateJobStatusParams{
		ID:         input.ID,
		Status:     input.Status,
		Attempts:   input.Attempts,
		LastError:  lastError,
		StartedAt:  pgtype.Timestamptz{},
		FinishedAt: pgtype.Timestamptz{},
	})
	if err != nil {
		return jobs.JobRecord{}, err
	}

	return toJobRecord(row), nil
}

func (r *JobRepository) GetByID(ctx context.Context, jobID uuid.UUID) (db.Job, error) {
	return r.queries.GetJobByID(ctx, jobID)
}

func toJobRecord(row db.Job) jobs.JobRecord {
	var lastError *string
	if row.LastError.Valid {
		lastError = &row.LastError.String
	}

	return jobs.JobRecord{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		RepositoryID:   row.RepositoryID,
		JobType:        row.JobType,
		Status:         row.Status,
		QueueName:      row.QueueName.String,
		Attempts:       row.Attempts,
		LastError:      lastError,
		Payload:        row.Payload,
	}
}
