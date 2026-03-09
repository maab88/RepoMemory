package auth

import "context"

type ExternalIdentity struct {
	Subject     string
	Issuer      string
	Email       string
	DisplayName string
	AvatarURL   string
}

type IdentityValidator interface {
	ValidateBearerToken(ctx context.Context, token string) (ExternalIdentity, error)
}

type IdentityMapper interface {
	MapToCurrentUser(ctx context.Context, identity ExternalIdentity) (CurrentUser, error)
}
