package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CurrentUser struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"displayName"`
	AvatarURL   string    `json:"avatarUrl,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

type contextKey string

const currentUserKey contextKey = "current_user"

func WithCurrentUser(ctx context.Context, user CurrentUser) context.Context {
	return context.WithValue(ctx, currentUserKey, user)
}

func CurrentUserFromContext(ctx context.Context) (CurrentUser, bool) {
	v := ctx.Value(currentUserKey)
	if v == nil {
		return CurrentUser{}, false
	}
	user, ok := v.(CurrentUser)
	return user, ok
}
