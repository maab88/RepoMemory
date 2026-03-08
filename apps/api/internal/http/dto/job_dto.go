package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	servicejobs "github.com/maab88/repomemory/apps/api/internal/services/jobs"
)

type JobDTO struct {
	ID         uuid.UUID       `json:"id"`
	JobType    string          `json:"jobType"`
	Status     string          `json:"status"`
	QueueName  string          `json:"queueName"`
	Attempts   int32           `json:"attempts"`
	LastError  *string         `json:"lastError"`
	Payload    json.RawMessage `json:"payload"`
	StartedAt  *time.Time      `json:"startedAt"`
	FinishedAt *time.Time      `json:"finishedAt"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
}

func ToJobDTO(job servicejobs.Job) JobDTO {
	return JobDTO{
		ID:         job.ID,
		JobType:    job.JobType,
		Status:     job.Status,
		QueueName:  job.QueueName,
		Attempts:   job.Attempts,
		LastError:  job.LastError,
		Payload:    job.Payload,
		StartedAt:  job.StartedAt,
		FinishedAt: job.FinishedAt,
		CreatedAt:  job.CreatedAt,
		UpdatedAt:  job.UpdatedAt,
	}
}
