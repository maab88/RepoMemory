package github

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type OAuthService struct {
	cfg         OAuthConfig
	state       StateService
	github      GitHubClient
	store       OAuthStore
	tokenSealer TokenSealer
}

func NewOAuthService(cfg OAuthConfig, state StateService, githubClient GitHubClient, store OAuthStore, tokenSealer TokenSealer) *OAuthService {
	return &OAuthService{
		cfg:         cfg,
		state:       state,
		github:      githubClient,
		store:       store,
		tokenSealer: tokenSealer,
	}
}

func (s *OAuthService) StartConnect(ctx context.Context, input OAuthStartInput) (string, error) {
	if !s.isConfigured() {
		return "", ErrOAuthNotConfigured
	}

	if input.OrganizationID != nil {
		allowed, err := s.store.UserHasMembership(ctx, input.UserID, *input.OrganizationID)
		if err != nil {
			return "", err
		}
		if !allowed {
			return "", ErrOrganizationAccessDenied
		}
	}

	stateValue, err := s.state.Generate(OAuthStatePayload{UserID: input.UserID, OrganizationID: input.OrganizationID})
	if err != nil {
		return "", err
	}

	authorizeURL, err := url.Parse(s.cfg.AuthorizeURL)
	if err != nil {
		return "", err
	}
	query := authorizeURL.Query()
	query.Set("client_id", s.cfg.ClientID)
	query.Set("redirect_uri", s.cfg.RedirectURL)
	query.Set("state", stateValue)
	query.Set("scope", s.cfg.Scope)
	authorizeURL.RawQuery = query.Encode()

	return authorizeURL.String(), nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, input OAuthCallbackInput) (GitHubConnectionSummary, error) {
	if !s.isConfigured() {
		return GitHubConnectionSummary{}, ErrOAuthNotConfigured
	}
	if strings.TrimSpace(input.Code) == "" {
		return GitHubConnectionSummary{}, ErrOAuthCodeMissing
	}
	if strings.TrimSpace(input.State) == "" {
		return GitHubConnectionSummary{}, ErrOAuthStateMissing
	}

	payload, err := s.state.Consume(input.State)
	if err != nil {
		if errors.Is(err, ErrStateExpired) {
			return GitHubConnectionSummary{}, err
		}
		return GitHubConnectionSummary{}, ErrInvalidState
	}
	if payload.UserID != input.UserID {
		return GitHubConnectionSummary{}, ErrStateUserMismatch
	}
	if payload.OrganizationID != nil {
		allowed, err := s.store.UserHasMembership(ctx, input.UserID, *payload.OrganizationID)
		if err != nil {
			return GitHubConnectionSummary{}, err
		}
		if !allowed {
			return GitHubConnectionSummary{}, ErrOrganizationAccessDenied
		}
	}

	token, err := s.github.ExchangeCode(ctx, input.Code, s.cfg.RedirectURL)
	if err != nil {
		return GitHubConnectionSummary{}, fmt.Errorf("%w", ErrTokenExchangeFailed)
	}

	viewer, err := s.github.GetViewer(ctx, token.AccessToken)
	if err != nil {
		return GitHubConnectionSummary{}, fmt.Errorf("%w", ErrGitHubUserFetchFailed)
	}

	sealedToken, err := s.tokenSealer.Seal(token.AccessToken)
	if err != nil {
		return GitHubConnectionSummary{}, err
	}

	return s.store.UpsertGitHubAccount(ctx, UpsertGitHubAccountInput{
		UserID:               input.UserID,
		GitHubUserID:         viewer.ID,
		GitHubLogin:          viewer.Login,
		AccessTokenEncrypted: sealedToken,
		TokenScope:           token.Scope,
	})
}

func (s *OAuthService) isConfigured() bool {
	return s.cfg.ClientID != "" && s.cfg.ClientSecret != "" && s.cfg.RedirectURL != "" && s.cfg.StateSecret != ""
}
