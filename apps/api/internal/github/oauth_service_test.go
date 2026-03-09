package github

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeStateService struct {
	stateValue string
	payload    OAuthStatePayload
	genErr     error
	consumeErr error
}

func (f *fakeStateService) Generate(_ OAuthStatePayload) (string, error) {
	if f.genErr != nil {
		return "", f.genErr
	}
	return f.stateValue, nil
}

func (f *fakeStateService) Consume(_ string) (OAuthStatePayload, error) {
	if f.consumeErr != nil {
		return OAuthStatePayload{}, f.consumeErr
	}
	return f.payload, nil
}

type fakeGitHubClient struct {
	token       GitHubToken
	viewer      GitHubUser
	exchangeErr error
	viewerErr   error
}

func (f *fakeGitHubClient) ExchangeCode(_ context.Context, _, _ string) (GitHubToken, error) {
	if f.exchangeErr != nil {
		return GitHubToken{}, f.exchangeErr
	}
	return f.token, nil
}

func (f *fakeGitHubClient) GetViewer(_ context.Context, _ string) (GitHubUser, error) {
	if f.viewerErr != nil {
		return GitHubUser{}, f.viewerErr
	}
	return f.viewer, nil
}

func (f *fakeGitHubClient) ListRepositories(_ context.Context, _ string) ([]GitHubRepository, error) {
	return nil, nil
}

type fakeTokenSealer struct {
	sealed string
	err    error
}

func (f fakeTokenSealer) Seal(_ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.sealed, nil
}

type fakeAccountStore struct {
	allowMembership bool
	upsertResult    GitHubConnectionSummary
	membershipErr   error
	upsertErr       error
	auditErr        error
	upsertCalls     int
	auditCalls      int
}

func (f *fakeAccountStore) UserHasMembership(_ context.Context, _, _ uuid.UUID) (bool, error) {
	if f.membershipErr != nil {
		return false, f.membershipErr
	}
	return f.allowMembership, nil
}

func (f *fakeAccountStore) UpsertGitHubAccount(_ context.Context, _ UpsertGitHubAccountInput) (GitHubConnectionSummary, error) {
	f.upsertCalls++
	if f.upsertErr != nil {
		return GitHubConnectionSummary{}, f.upsertErr
	}
	return f.upsertResult, nil
}

func (f *fakeAccountStore) InsertGitHubConnectionAuditLog(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _ GitHubConnectionSummary) error {
	f.auditCalls++
	return f.auditErr
}

func newConfiguredOAuthService(state StateService, githubClient GitHubClient, store OAuthStore) *OAuthService {
	return NewOAuthService(OAuthConfig{
		ClientID:     "cid",
		ClientSecret: "secret",
		AuthorizeURL: "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		APIBaseURL:   "https://api.github.com",
		RedirectURL:  "http://localhost:3000/integrations/github/callback",
		StateSecret:  "state-secret",
		StateTTL:     10 * time.Minute,
		Scope:        "repo read:user",
	}, state, githubClient, store, fakeTokenSealer{sealed: "encrypted-token"})
}

func TestStartConnectReturnsRedirectURL(t *testing.T) {
	userID := uuid.New()
	service := newConfiguredOAuthService(&fakeStateService{stateValue: "state123"}, &fakeGitHubClient{}, &fakeAccountStore{allowMembership: true})

	url, err := service.StartConnect(context.Background(), OAuthStartInput{UserID: userID})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if url == "" {
		t.Fatal("expected redirect url")
	}
}

func TestHandleCallbackInvalidStateFails(t *testing.T) {
	userID := uuid.New()
	service := newConfiguredOAuthService(&fakeStateService{consumeErr: ErrInvalidState}, &fakeGitHubClient{}, &fakeAccountStore{allowMembership: true})

	_, err := service.HandleCallback(context.Background(), OAuthCallbackInput{UserID: userID, Code: "abc", State: "bad"})
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("expected invalid state, got %v", err)
	}
}

func TestHandleCallbackTokenExchangeFailure(t *testing.T) {
	userID := uuid.New()
	service := newConfiguredOAuthService(
		&fakeStateService{payload: OAuthStatePayload{UserID: userID}},
		&fakeGitHubClient{exchangeErr: errors.New("bad token")},
		&fakeAccountStore{allowMembership: true},
	)

	_, err := service.HandleCallback(context.Background(), OAuthCallbackInput{UserID: userID, Code: "abc", State: "state"})
	if !errors.Is(err, ErrTokenExchangeFailed) {
		t.Fatalf("expected token exchange error, got %v", err)
	}
}

func TestHandleCallbackUserFetchFailure(t *testing.T) {
	userID := uuid.New()
	service := newConfiguredOAuthService(
		&fakeStateService{payload: OAuthStatePayload{UserID: userID}},
		&fakeGitHubClient{token: GitHubToken{AccessToken: "token"}, viewerErr: errors.New("no user")},
		&fakeAccountStore{allowMembership: true},
	)

	_, err := service.HandleCallback(context.Background(), OAuthCallbackInput{UserID: userID, Code: "abc", State: "state"})
	if !errors.Is(err, ErrGitHubUserFetchFailed) {
		t.Fatalf("expected github user fetch error, got %v", err)
	}
}

func TestHandleCallbackUpsertExistingAndNewCases(t *testing.T) {
	userID := uuid.New()
	baseState := &fakeStateService{payload: OAuthStatePayload{UserID: userID}}
	ghClient := &fakeGitHubClient{token: GitHubToken{AccessToken: "token", Scope: "repo"}, viewer: GitHubUser{ID: 1, Login: "octocat"}}

	newStore := &fakeAccountStore{allowMembership: true, upsertResult: GitHubConnectionSummary{ID: uuid.New(), GitHubLogin: "octocat", GitHubUserID: "1"}}
	service := newConfiguredOAuthService(baseState, ghClient, newStore)
	if _, err := service.HandleCallback(context.Background(), OAuthCallbackInput{UserID: userID, Code: "a", State: "b"}); err != nil {
		t.Fatalf("unexpected error for new account: %v", err)
	}

	existingStore := &fakeAccountStore{allowMembership: true, upsertResult: GitHubConnectionSummary{ID: uuid.New(), GitHubLogin: "octocat", GitHubUserID: "1"}}
	service2 := newConfiguredOAuthService(baseState, ghClient, existingStore)
	if _, err := service2.HandleCallback(context.Background(), OAuthCallbackInput{UserID: userID, Code: "a", State: "b"}); err != nil {
		t.Fatalf("unexpected error for existing account: %v", err)
	}

	if newStore.upsertCalls != 1 || existingStore.upsertCalls != 1 {
		t.Fatalf("expected upsert in both scenarios, got new=%d existing=%d", newStore.upsertCalls, existingStore.upsertCalls)
	}
	if newStore.auditCalls != 1 || existingStore.auditCalls != 1 {
		t.Fatalf("expected audit insert in both scenarios, got new=%d existing=%d", newStore.auditCalls, existingStore.auditCalls)
	}
}
