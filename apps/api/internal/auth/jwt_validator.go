package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidAuthToken = errors.New("invalid auth token")
)

type JWTValidatorConfig struct {
	Secret   string
	Issuer   string
	Audience string
}

type JWTIdentityValidator struct {
	secret   []byte
	issuer   string
	audience string
}

type authClaims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	jwt.RegisteredClaims
}

func NewJWTIdentityValidator(cfg JWTValidatorConfig) (*JWTIdentityValidator, error) {
	if strings.TrimSpace(cfg.Secret) == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	if strings.TrimSpace(cfg.Issuer) == "" {
		return nil, fmt.Errorf("jwt issuer is required")
	}
	if strings.TrimSpace(cfg.Audience) == "" {
		return nil, fmt.Errorf("jwt audience is required")
	}
	return &JWTIdentityValidator{
		secret:   []byte(cfg.Secret),
		issuer:   cfg.Issuer,
		audience: cfg.Audience,
	}, nil
}

func (v *JWTIdentityValidator) ValidateBearerToken(_ context.Context, token string) (ExternalIdentity, error) {
	claims := &authClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		method, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, ErrInvalidAuthToken
		}
		return v.secret, nil
	}, jwt.WithIssuer(v.issuer), jwt.WithAudience(v.audience), jwt.WithLeeway(10*time.Second))
	if err != nil || !parsed.Valid {
		return ExternalIdentity{}, ErrInvalidAuthToken
	}

	subject := strings.TrimSpace(claims.Subject)
	issuer := strings.TrimSpace(claims.Issuer)
	if subject == "" || issuer == "" {
		return ExternalIdentity{}, ErrInvalidAuthToken
	}

	return ExternalIdentity{
		Subject:     subject,
		Issuer:      issuer,
		Email:       strings.TrimSpace(claims.Email),
		DisplayName: strings.TrimSpace(claims.Name),
		AvatarURL:   strings.TrimSpace(claims.Picture),
	}, nil
}
