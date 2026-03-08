package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type CreateJobInput struct {
	OrganizationID *uuid.UUID
	RepositoryID   *uuid.UUID
	JobType        string
	QueueName      string
	Payload        []byte
}

type JobRecord struct {
	ID             uuid.UUID
	OrganizationID *uuid.UUID
	RepositoryID   *uuid.UUID
	JobType        string
	Status         string
	QueueName      string
	Attempts       int32
	LastError      *string
	Payload        []byte
}

type JobRepository interface {
	CreateJob(ctx context.Context, input CreateJobInput) (JobRecord, error)
	UpdateJobLifecycle(ctx context.Context, input UpdateJobLifecycleInput) (JobRecord, error)
}

type UpdateJobLifecycleInput struct {
	ID        uuid.UUID
	Status    string
	Attempts  int32
	LastError *string
}

type AsynqTaskEnqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type Enqueuer struct {
	jobRepository JobRepository
	client        AsynqTaskEnqueuer
}

func NewEnqueuer(jobRepository JobRepository, client AsynqTaskEnqueuer) *Enqueuer {
	return &Enqueuer{jobRepository: jobRepository, client: client}
}

type repoTaskEnvelope[T any] struct {
	JobID   uuid.UUID `json:"jobId"`
	Payload T         `json:"payload"`
}

func (e *Enqueuer) EnqueueRepoInitialSync(ctx context.Context, payload RepoInitialSyncPayload) (JobRecord, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return JobRecord{}, fmt.Errorf("marshal job payload: %w", err)
	}

	job, err := e.jobRepository.CreateJob(ctx, CreateJobInput{
		OrganizationID: &payload.OrganizationID,
		RepositoryID:   &payload.RepositoryID,
		JobType:        TaskRepoInitialSync,
		QueueName:      QueueDefault,
		Payload:        payloadJSON,
	})
	if err != nil {
		return JobRecord{}, err
	}

	taskPayload, err := json.Marshal(repoTaskEnvelope[RepoInitialSyncPayload]{
		JobID:   job.ID,
		Payload: payload,
	})
	if err != nil {
		enqueueErr := fmt.Sprintf("marshal asynq task payload: %v", err)
		_, _ = e.jobRepository.UpdateJobLifecycle(ctx, UpdateJobLifecycleInput{
			ID:        job.ID,
			Status:    StatusFailed,
			Attempts:  job.Attempts,
			LastError: &enqueueErr,
		})
		return JobRecord{}, fmt.Errorf("marshal asynq task payload: %w", err)
	}

	task := asynq.NewTask(TaskRepoInitialSync, taskPayload)
	if _, err := e.client.Enqueue(
		task,
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(5),
	); err != nil {
		enqueueErr := fmt.Sprintf("enqueue task: %v", err)
		_, _ = e.jobRepository.UpdateJobLifecycle(ctx, UpdateJobLifecycleInput{
			ID:        job.ID,
			Status:    StatusFailed,
			Attempts:  job.Attempts,
			LastError: &enqueueErr,
		})
		return JobRecord{}, fmt.Errorf("enqueue repo initial sync task: %w", err)
	}

	return job, nil
}
