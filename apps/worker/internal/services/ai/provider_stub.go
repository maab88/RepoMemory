package ai

import (
	"context"
	"sync"
)

type StubProvider struct {
	mu       sync.Mutex
	response string
	err      error
}

func NewStubProvider() *StubProvider {
	return &StubProvider{}
}

func (p *StubProvider) Name() string { return ProviderStub }

func (p *StubProvider) SetResult(response string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.response = response
	p.err = err
}

func (p *StubProvider) CompleteJSON(context.Context, CompletionRequest) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.response, p.err
}
