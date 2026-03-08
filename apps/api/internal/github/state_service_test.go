package github

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMemoryStateServiceGenerateAndConsume(t *testing.T) {
	svc := NewMemoryStateService("secret", time.Minute)
	userID := uuid.New()
	orgID := uuid.New()

	state, err := svc.Generate(OAuthStatePayload{UserID: userID, OrganizationID: &orgID})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if state == "" {
		t.Fatal("expected non-empty state")
	}

	payload, err := svc.Consume(state)
	if err != nil {
		t.Fatalf("consume failed: %v", err)
	}
	if payload.UserID != userID {
		t.Fatalf("unexpected user id: %s", payload.UserID)
	}
	if payload.OrganizationID == nil || *payload.OrganizationID != orgID {
		t.Fatalf("unexpected organization id payload: %+v", payload.OrganizationID)
	}
}

func TestMemoryStateServiceInvalidStateFails(t *testing.T) {
	svc := NewMemoryStateService("secret", time.Minute)
	if _, err := svc.Consume("bad-state"); err == nil {
		t.Fatal("expected error for invalid state")
	}
}

func TestMemoryStateServiceSingleUse(t *testing.T) {
	svc := NewMemoryStateService("secret", time.Minute)
	state, err := svc.Generate(OAuthStatePayload{UserID: uuid.New()})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if _, err := svc.Consume(state); err != nil {
		t.Fatalf("first consume should pass: %v", err)
	}
	if _, err := svc.Consume(state); err == nil {
		t.Fatal("expected second consume to fail")
	}
}
