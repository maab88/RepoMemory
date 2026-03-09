package github

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidState             = errors.New("invalid oauth state")
	ErrStateExpired             = errors.New("oauth state expired")
	ErrStateUserMismatch        = errors.New("oauth state does not belong to current user")
	ErrOAuthCodeMissing         = errors.New("oauth code is required")
	ErrOAuthStateMissing        = errors.New("oauth state is required")
	ErrTokenExchangeFailed      = errors.New("oauth token exchange failed")
	ErrGitHubUserFetchFailed    = errors.New("failed to fetch github user")
	ErrOrganizationAccessDenied = errors.New("organization access denied")
	ErrOAuthNotConfigured       = errors.New("github oauth is not configured")
	ErrGitHubNotConnected       = errors.New("github account is not connected")
	ErrGitHubReconnectRequired  = errors.New("github reconnect required")
	ErrGitHubRateLimited        = errors.New("github api rate limited")
	ErrImportRepositoriesEmpty  = errors.New("repositories payload is empty")
	ErrInvalidRepositoryPayload = errors.New("invalid repository payload")
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	AuthorizeURL string
	TokenURL     string
	APIBaseURL   string
	RedirectURL  string
	StateSecret  string
	StateTTL     time.Duration
	Scope        string
}

type OAuthStartInput struct {
	UserID         uuid.UUID
	OrganizationID *uuid.UUID
}

type OAuthCallbackInput struct {
	UserID uuid.UUID
	Code   string
	State  string
}

type GitHubConnectionSummary struct {
	ID           uuid.UUID `json:"id"`
	GitHubLogin  string    `json:"githubLogin"`
	GitHubUserID string    `json:"githubUserId"`
	ConnectedAt  time.Time `json:"connectedAt"`
}

type OAuthStatePayload struct {
	UserID         uuid.UUID
	OrganizationID *uuid.UUID
}

type StateService interface {
	Generate(input OAuthStatePayload) (string, error)
	Consume(state string) (OAuthStatePayload, error)
}

type GitHubClient interface {
	ExchangeCode(ctx context.Context, code, redirectURL string) (GitHubToken, error)
	GetViewer(ctx context.Context, accessToken string) (GitHubUser, error)
	ListRepositories(ctx context.Context, accessToken string) ([]GitHubRepository, error)
}

type TokenSealer interface {
	Seal(token string) (string, error)
}

type OAuthStore interface {
	UpsertGitHubAccount(ctx context.Context, input UpsertGitHubAccountInput) (GitHubConnectionSummary, error)
	UserHasMembership(ctx context.Context, userID, organizationID uuid.UUID) (bool, error)
}

type RepositoryStore interface {
	UserHasMembership(ctx context.Context, userID, organizationID uuid.UUID) (bool, error)
	GetLatestGitHubAccountForUser(ctx context.Context, userID uuid.UUID) (ConnectedGitHubAccount, error)
	UpsertRepositoryForOrganization(ctx context.Context, organizationID uuid.UUID, repo GitHubRepository) (ImportedRepository, error)
	EnsureRepositorySyncState(ctx context.Context, repositoryID uuid.UUID) error
	InsertRepositoryImportAuditLog(ctx context.Context, actorUserID, organizationID, repositoryID uuid.UUID, repo GitHubRepository) error
}

type UpsertGitHubAccountInput struct {
	UserID               uuid.UUID
	GitHubUserID         int64
	GitHubLogin          string
	AccessTokenEncrypted string
	TokenScope           string
}

type GitHubToken struct {
	AccessToken string
	Scope       string
}

type GitHubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}

type GitHubRepository struct {
	GitHubRepoID  int64  `json:"githubRepoId"`
	OwnerLogin    string `json:"ownerLogin"`
	Name          string `json:"name"`
	FullName      string `json:"fullName"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"defaultBranch"`
	HTMLURL       string `json:"htmlUrl"`
	Description   string `json:"description,omitempty"`
}

type ConnectedGitHubAccount struct {
	UserID               uuid.UUID
	AccessTokenEncrypted string
}

type ImportRepositoriesInput struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	Repositories   []GitHubRepository
}

type ImportedRepository struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organizationId"`
	GitHubRepoID   string    `json:"githubRepoId"`
	OwnerLogin     string    `json:"ownerLogin"`
	Name           string    `json:"name"`
	FullName       string    `json:"fullName"`
	Private        bool      `json:"private"`
	DefaultBranch  string    `json:"defaultBranch"`
	HTMLURL        string    `json:"htmlUrl"`
	Description    string    `json:"description,omitempty"`
	ImportedAt     time.Time `json:"importedAt"`
}

type PlaintextTokenSealer struct{}

func (PlaintextTokenSealer) Seal(token string) (string, error) {
	if token == "" {
		return "", ErrTokenExchangeFailed
	}
	return token, nil
}
