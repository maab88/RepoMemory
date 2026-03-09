package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPGitHubClient struct {
	client       *http.Client
	clientID     string
	clientSecret string
	tokenURL     string
	apiBaseURL   string
}

func NewHTTPGitHubClient(clientID, clientSecret, tokenURL, apiBaseURL string) *HTTPGitHubClient {
	if tokenURL == "" {
		tokenURL = "https://github.com/login/oauth/access_token"
	}
	if apiBaseURL == "" {
		apiBaseURL = "https://api.github.com"
	}
	return &HTTPGitHubClient{
		client:       &http.Client{Timeout: 10 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
		apiBaseURL:   strings.TrimRight(apiBaseURL, "/"),
	}
}

type tokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

type githubAPIError struct {
	Message string `json:"message"`
}

func (c *HTTPGitHubClient) ExchangeCode(ctx context.Context, code, redirectURL string) (GitHubToken, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return GitHubToken{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "RepoMemory")

	resp, err := c.client.Do(req)
	if err != nil {
		return GitHubToken{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return GitHubToken{}, err
	}
	if resp.StatusCode >= 400 {
		return GitHubToken{}, classifyGitHubHTTPError(resp.StatusCode, resp.Header, body)
	}

	var parsed tokenExchangeResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return GitHubToken{}, err
	}
	if parsed.Error != "" || parsed.AccessToken == "" {
		return GitHubToken{}, ErrTokenExchangeFailed
	}

	return GitHubToken{AccessToken: parsed.AccessToken, Scope: parsed.Scope}, nil
}

func (c *HTTPGitHubClient) GetViewer(ctx context.Context, accessToken string) (GitHubUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+"/user", nil)
	if err != nil {
		return GitHubUser{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "RepoMemory")

	resp, err := c.client.Do(req)
	if err != nil {
		return GitHubUser{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return GitHubUser{}, classifyGitHubHTTPError(resp.StatusCode, resp.Header, body)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return GitHubUser{}, err
	}
	if user.ID == 0 || user.Login == "" {
		return GitHubUser{}, fmt.Errorf("github user response missing fields")
	}

	return user, nil
}

func (c *HTTPGitHubClient) ListRepositories(ctx context.Context, accessToken string) ([]GitHubRepository, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiBaseURL+"/user/repos?per_page=100&sort=updated&affiliation=owner,collaborator,organization_member", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "RepoMemory")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, classifyGitHubHTTPError(resp.StatusCode, resp.Header, body)
	}

	var payload []struct {
		ID            int64   `json:"id"`
		Name          string  `json:"name"`
		FullName      string  `json:"full_name"`
		Private       bool    `json:"private"`
		DefaultBranch string  `json:"default_branch"`
		HTMLURL       string  `json:"html_url"`
		Description   *string `json:"description"`
		Owner         struct {
			Login string `json:"login"`
		} `json:"owner"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	out := make([]GitHubRepository, 0, len(payload))
	for _, item := range payload {
		description := ""
		if item.Description != nil {
			description = *item.Description
		}
		if item.ID == 0 || item.Name == "" || item.FullName == "" || item.Owner.Login == "" || item.HTMLURL == "" || item.DefaultBranch == "" {
			continue
		}
		out = append(out, GitHubRepository{
			GitHubRepoID:  item.ID,
			OwnerLogin:    item.Owner.Login,
			Name:          item.Name,
			FullName:      item.FullName,
			Private:       item.Private,
			DefaultBranch: item.DefaultBranch,
			HTMLURL:       item.HTMLURL,
			Description:   description,
		})
	}

	return out, nil
}

var _ GitHubClient = (*HTTPGitHubClient)(nil)

func classifyGitHubHTTPError(status int, headers http.Header, body []byte) error {
	if status == http.StatusUnauthorized {
		return ErrGitHubReconnectRequired
	}
	if status == http.StatusForbidden && isRateLimited(headers, body) {
		return ErrGitHubRateLimited
	}
	if status == http.StatusTooManyRequests {
		return ErrGitHubRateLimited
	}
	return fmt.Errorf("github api error status=%d", status)
}

func isRateLimited(headers http.Header, body []byte) bool {
	if headers.Get("X-RateLimit-Remaining") == "0" {
		return true
	}
	var parsed githubAPIError
	if err := json.Unmarshal(body, &parsed); err == nil {
		message := strings.ToLower(strings.TrimSpace(parsed.Message))
		if strings.Contains(message, "rate limit") {
			return true
		}
	}
	return false
}
