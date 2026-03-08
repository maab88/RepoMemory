package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type fakeJobRepository struct {
	created JobRecord
	updated UpdateJobLifecycleInput
}

func (f *fakeJobRepository) CreateJob(_ context.Context, input CreateJobInput) (JobRecord, error) {
	f.created = JobRecord{
		ID:             uuid.New(),
		OrganizationID: input.OrganizationID,
		RepositoryID:   input.RepositoryID,
		JobType:        input.JobType,
		Status:         StatusQueued,
		QueueName:      input.QueueName,
		Attempts:       0,
		Payload:        input.Payload,
	}
	return f.created, nil
}

func (f *fakeJobRepository) UpdateJobLifecycle(_ context.Context, input UpdateJobLifecycleInput) (JobRecord, error) {
	f.updated = input
	return JobRecord{}, nil
}

type fakeAsynqClient struct {
	task *asynq.Task
	err  error
}

func (f *fakeAsynqClient) Enqueue(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
	f.task = task
	if f.err != nil {
		return nil, f.err
	}
	return &asynq.TaskInfo{}, nil
}

func TestEnqueueRepoInitialSync(t *testing.T) {
	repoID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()
	jobRepo := &fakeJobRepository{}
	client := &fakeAsynqClient{}
	enqueuer := NewEnqueuer(jobRepo, client)

	job, err := enqueuer.EnqueueRepoInitialSync(context.Background(), RepoInitialSyncPayload{
		RepositoryID:      repoID,
		OrganizationID:    orgID,
		TriggeredByUserID: userID,
	})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	if job.Status != StatusQueued {
		t.Fatalf("expected queued status, got %s", job.Status)
	}
	if client.task == nil {
		t.Fatal("expected asynq task to be enqueued")
	}

	var payload map[string]any
	if err := json.Unmarshal(client.task.Payload(), &payload); err != nil {
		t.Fatalf("parse asynq payload: %v", err)
	}
	if payload["jobId"] == "" {
		t.Fatal("expected jobId in task payload")
	}
}

func TestEnqueueRepoInitialSyncFailureMarksJobFailed(t *testing.T) {
	jobRepo := &fakeJobRepository{}
	client := &fakeAsynqClient{err: errors.New("redis unavailable")}
	enqueuer := NewEnqueuer(jobRepo, client)

	_, err := enqueuer.EnqueueRepoInitialSync(context.Background(), RepoInitialSyncPayload{
		RepositoryID:      uuid.New(),
		OrganizationID:    uuid.New(),
		TriggeredByUserID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected enqueue error")
	}
	if jobRepo.updated.Status != StatusFailed {
		t.Fatalf("expected failed status update, got %s", jobRepo.updated.Status)
	}
}

func TestEnqueueRepoGenerateMemory(t *testing.T) {
	repoID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()
	jobRepo := &fakeJobRepository{}
	client := &fakeAsynqClient{}
	enqueuer := NewEnqueuer(jobRepo, client)

	job, err := enqueuer.EnqueueRepoGenerateMemory(context.Background(), RepoGenerateMemoryPayload{
		RepositoryID:      repoID,
		OrganizationID:    orgID,
		TriggeredByUserID: userID,
	})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	if job.JobType != TaskRepoGenerateMemory {
		t.Fatalf("expected job type %s, got %s", TaskRepoGenerateMemory, job.JobType)
	}
	if client.task == nil {
		t.Fatal("expected asynq task to be enqueued")
	}
	if client.task.Type() != TaskRepoGenerateMemory {
		t.Fatalf("expected task type %s, got %s", TaskRepoGenerateMemory, client.task.Type())
	}
}

func TestEnqueueRepoGenerateMemoryFailureMarksJobFailed(t *testing.T) {
	jobRepo := &fakeJobRepository{}
	client := &fakeAsynqClient{err: errors.New("redis unavailable")}
	enqueuer := NewEnqueuer(jobRepo, client)

	_, err := enqueuer.EnqueueRepoGenerateMemory(context.Background(), RepoGenerateMemoryPayload{
		RepositoryID:      uuid.New(),
		OrganizationID:    uuid.New(),
		TriggeredByUserID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected enqueue error")
	}
	if jobRepo.updated.Status != StatusFailed {
		t.Fatalf("expected failed status update, got %s", jobRepo.updated.Status)
	}
}

func TestEnqueueRepoGenerateDigest(t *testing.T) {
	repoID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()
	jobRepo := &fakeJobRepository{}
	client := &fakeAsynqClient{}
	enqueuer := NewEnqueuer(jobRepo, client)

	job, err := enqueuer.EnqueueRepoGenerateDigest(context.Background(), RepoGenerateDigestPayload{
		RepositoryID:      repoID,
		OrganizationID:    orgID,
		TriggeredByUserID: userID,
	})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	if job.JobType != TaskRepoGenerateDigest {
		t.Fatalf("expected job type %s, got %s", TaskRepoGenerateDigest, job.JobType)
	}
	if client.task == nil {
		t.Fatal("expected asynq task to be enqueued")
	}
	if client.task.Type() != TaskRepoGenerateDigest {
		t.Fatalf("expected task type %s, got %s", TaskRepoGenerateDigest, client.task.Type())
	}
}

func TestEnqueueRepoGenerateDigestFailureMarksJobFailed(t *testing.T) {
	jobRepo := &fakeJobRepository{}
	client := &fakeAsynqClient{err: errors.New("redis unavailable")}
	enqueuer := NewEnqueuer(jobRepo, client)

	_, err := enqueuer.EnqueueRepoGenerateDigest(context.Background(), RepoGenerateDigestPayload{
		RepositoryID:      uuid.New(),
		OrganizationID:    uuid.New(),
		TriggeredByUserID: uuid.New(),
	})
	if err == nil {
		t.Fatal("expected enqueue error")
	}
	if jobRepo.updated.Status != StatusFailed {
		t.Fatalf("expected failed status update, got %s", jobRepo.updated.Status)
	}
}
