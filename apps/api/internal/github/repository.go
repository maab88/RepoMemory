package github

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
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

var _ AccountStore = (*AccountRepository)(nil)
