package ai

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrProviderDisabled = errors.New("ai provider disabled")
	ErrInvalidResponse  = errors.New("ai returned invalid response")
)

type CompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
}

type Provider interface {
	Name() string
	CompleteJSON(ctx context.Context, request CompletionRequest) (string, error)
}

type disabledProvider struct{}

func (p *disabledProvider) Name() string { return ProviderDisabled }

func (p *disabledProvider) CompleteJSON(context.Context, CompletionRequest) (string, error) {
	return "", ErrProviderDisabled
}

func NewProvider(config Config) (Provider, error) {
	switch NormalizeProvider(config.Provider) {
	case ProviderStub:
		return NewStubProvider(), nil
	case ProviderOpenAI:
		return NewOpenAIProvider(config)
	default:
		return &disabledProvider{}, nil
	}
}

func mustPrompt(value, name string) error {
	if value == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}
