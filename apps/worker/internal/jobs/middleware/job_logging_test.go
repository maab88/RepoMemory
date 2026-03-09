package middleware

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestParseEnvelopeExtractsTraceFields(t *testing.T) {
	payload := taskEnvelope{
		JobID: uuid.New(),
		Payload: taskPayloadMeta{
			RepositoryID:      uuid.New(),
			OrganizationID:    uuid.New(),
			TriggeredByUserID: uuid.New(),
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	meta := parseEnvelope(raw)
	if meta.jobID == "" || meta.repositoryID == "" || meta.organizationID == "" || meta.triggeredByUserID == "" {
		t.Fatalf("expected all trace fields, got %+v", meta)
	}
}

func TestParseEnvelopeHandlesInvalidJSON(t *testing.T) {
	meta := parseEnvelope([]byte("{bad-json"))
	if meta.jobID != "" || meta.repositoryID != "" || meta.organizationID != "" || meta.triggeredByUserID != "" {
		t.Fatalf("expected empty metadata, got %+v", meta)
	}
}
