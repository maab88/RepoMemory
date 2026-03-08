package ai

import (
	"context"
	"errors"
	"testing"
)

func TestNewProviderDefaultsToDisabled(t *testing.T) {
	p, err := NewProvider(Config{})
	if err != nil {
		t.Fatalf("NewProvider error: %v", err)
	}
	if p.Name() != ProviderDisabled {
		t.Fatalf("expected disabled provider, got %s", p.Name())
	}
	if _, err := p.CompleteJSON(context.Background(), CompletionRequest{}); !errors.Is(err, ErrProviderDisabled) {
		t.Fatalf("expected disabled error, got %v", err)
	}
}

func TestStubProvider(t *testing.T) {
	p, err := NewProvider(Config{Provider: ProviderStub})
	if err != nil {
		t.Fatalf("NewProvider error: %v", err)
	}
	stub := p.(*StubProvider)
	stub.SetResult(`{"ok":true}`, nil)
	got, err := stub.CompleteJSON(context.Background(), CompletionRequest{SystemPrompt: "s", UserPrompt: "u"})
	if err != nil {
		t.Fatalf("CompleteJSON error: %v", err)
	}
	if got == "" {
		t.Fatal("expected stub response")
	}
}

func TestOpenAIProviderRequiresKey(t *testing.T) {
	_, err := NewProvider(Config{Provider: ProviderOpenAI})
	if err == nil {
		t.Fatal("expected missing key error")
	}
}
