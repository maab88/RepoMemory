package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OpenAIProvider struct {
	apiKey string
	url    string
	model  string
	client *http.Client
}

func NewOpenAIProvider(config Config) (*OpenAIProvider, error) {
	if strings.TrimSpace(config.OpenAIKey) == "" {
		return nil, fmt.Errorf("openai provider: missing api key")
	}
	url := strings.TrimSpace(config.OpenAIURL)
	if url == "" {
		url = "https://api.openai.com/v1/chat/completions"
	}
	model := strings.TrimSpace(config.OpenAIModel)
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{
		apiKey: config.OpenAIKey,
		url:    url,
		model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (p *OpenAIProvider) Name() string { return ProviderOpenAI }

type openAIChatRequest struct {
	Model          string              `json:"model"`
	ResponseFormat map[string]string   `json:"response_format"`
	Messages       []openAIChatMessage `json:"messages"`
	Temperature    float64             `json:"temperature"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message openAIChatMessage `json:"message"`
	} `json:"choices"`
}

func (p *OpenAIProvider) CompleteJSON(ctx context.Context, request CompletionRequest) (string, error) {
	if err := mustPrompt(request.SystemPrompt, "system prompt"); err != nil {
		return "", err
	}
	if err := mustPrompt(request.UserPrompt, "user prompt"); err != nil {
		return "", err
	}

	body, err := json.Marshal(openAIChatRequest{
		Model: p.model,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Messages: []openAIChatMessage{
			{Role: "system", Content: request.SystemPrompt},
			{Role: "user", Content: request.UserPrompt},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("openai request failed with status %d", resp.StatusCode)
	}

	var parsed openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", ErrInvalidResponse
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", ErrInvalidResponse
	}
	return content, nil
}
