package github

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type AccountRepository struct {
	queries *db.Queries
}

func NewAccountRepository(queries *db.Queries) *AccountRepository {
	return &AccountRepository{queries: queries}
}

func (r *AccountRepository) UserHasMembership(ctx context.Context, userID, organizationID uuid.UUID) (bool, error) {
	return r.queries.UserHasMembership(ctx, db.UserHasMembershipParams{UserID: userID, OrganizationID: organizationID})
}

func (r *AccountRepository) UpsertGitHubAccount(ctx context.Context, input UpsertGitHubAccountInput) (GitHubConnectionSummary, error) {
	scope := pgtype.Text{}
	if input.TokenScope != "" {
		scope = pgtype.Text{String: input.TokenScope, Valid: true}
	}

	row, err := r.queries.UpsertGithubAccount(ctx, db.UpsertGithubAccountParams{
		UserID:               input.UserID,
		GithubUserID:         input.GitHubUserID,
		GithubLogin:          input.GitHubLogin,
		AccessTokenEncrypted: input.AccessTokenEncrypted,
		TokenScope:           scope,
	})
	if err != nil {
		return GitHubConnectionSummary{}, err
	}

	connectedAt := time.Time{}
	if row.ConnectedAt.Valid {
		connectedAt = row.ConnectedAt.Time.UTC()
	}

	return GitHubConnectionSummary{
		ID:           row.ID,
		GitHubLogin:  row.GithubLogin,
		GitHubUserID: strconv.FormatInt(row.GithubUserID, 10),
		ConnectedAt:  connectedAt,
	}, nil
}

func (r *AccountRepository) GetLatestGitHubAccountForUser(ctx context.Context, userID uuid.UUID) (ConnectedGitHubAccount, error) {
	row, err := r.queries.GetLatestGithubAccountForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ConnectedGitHubAccount{}, ErrGitHubNotConnected
		}
		return ConnectedGitHubAccount{}, err
	}
	return ConnectedGitHubAccount{
		UserID:               row.UserID,
		AccessTokenEncrypted: row.AccessTokenEncrypted,
	}, nil
}

func (r *AccountRepository) UpsertRepositoryForOrganization(ctx context.Context, organizationID uuid.UUID, repo GitHubRepository) (ImportedRepository, error) {
	description := pgtype.Text{}
	if repo.Description != "" {
		description = pgtype.Text{String: repo.Description, Valid: true}
	}

	row, err := r.queries.UpsertRepository(ctx, db.UpsertRepositoryParams{
		OrganizationID: organizationID,
		GithubRepoID:   repo.GitHubRepoID,
		OwnerLogin:     repo.OwnerLogin,
		Name:           repo.Name,
		FullName:       repo.FullName,
		Private:        repo.Private,
		DefaultBranch:  repo.DefaultBranch,
		HtmlUrl:        repo.HTMLURL,
		Description:    description,
		IsActive:       true,
	})
	if err != nil {
		return ImportedRepository{}, err
	}

	importedAt := time.Time{}
	if row.ImportedAt.Valid {
		importedAt = row.ImportedAt.Time.UTC()
	}

	return ImportedRepository{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		GitHubRepoID:   strconv.FormatInt(row.GithubRepoID, 10),
		OwnerLogin:     row.OwnerLogin,
		Name:           row.Name,
		FullName:       row.FullName,
		Private:        row.Private,
		DefaultBranch:  row.DefaultBranch,
		HTMLURL:        row.HtmlUrl,
		Description:    row.Description.String,
		ImportedAt:     importedAt,
	}, nil
}

func (r *AccountRepository) EnsureRepositorySyncState(ctx context.Context, repositoryID uuid.UUID) error {
	_, err := r.queries.UpsertRepositorySyncState(ctx, db.UpsertRepositorySyncStateParams{
		RepositoryID:         repositoryID,
		LastPrSyncAt:         pgtype.Timestamptz{},
		LastIssueSyncAt:      pgtype.Timestamptz{},
		LastSuccessfulSyncAt: pgtype.Timestamptz{},
		LastSyncStatus:       pgtype.Text{String: "pending_import", Valid: true},
		LastSyncError:        pgtype.Text{},
	})
	return err
}

func (r *AccountRepository) InsertRepositoryImportAuditLog(ctx context.Context, actorUserID, organizationID, repositoryID uuid.UUID, repo GitHubRepository) error {
	metadata, err := json.Marshal(map[string]any{
		"githubRepoId": repo.GitHubRepoID,
		"fullName":     repo.FullName,
		"ownerLogin":   repo.OwnerLogin,
	})
	if err != nil {
		return err
	}
	_, err = r.queries.InsertAuditLog(ctx, db.InsertAuditLogParams{
		OrganizationID: &organizationID,
		ActorUserID:    &actorUserID,
		Action:         "repository.imported",
		EntityType:     "repository",
		EntityID:       &repositoryID,
		Metadata:       metadata,
	})
	return err
}

func (r *AccountRepository) InsertGitHubConnectionAuditLog(ctx context.Context, actorUserID uuid.UUID, organizationID *uuid.UUID, account GitHubConnectionSummary) error {
	metadata, err := json.Marshal(map[string]any{
		"githubLogin":  account.GitHubLogin,
		"githubUserId": account.GitHubUserID,
	})
	if err != nil {
		return err
	}
	_, err = r.queries.InsertAuditLog(ctx, db.InsertAuditLogParams{
		OrganizationID: organizationID,
		ActorUserID:    &actorUserID,
		Action:         "github.connection_succeeded",
		EntityType:     "github_account",
		EntityID:       &account.ID,
		Metadata:       metadata,
	})
	return err
}

var _ OAuthStore = (*AccountRepository)(nil)
var _ RepositoryStore = (*AccountRepository)(nil)
