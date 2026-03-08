package github

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

type stateEnvelope struct {
	Nonce          string  `json:"nonce"`
	UserID         string  `json:"userId"`
	OrganizationID *string `json:"organizationId,omitempty"`
	ExpiresAt      int64   `json:"exp"`
}

type MemoryStateService struct {
	secret  []byte
	ttl     time.Duration
	now     func() time.Time
	mu      sync.Mutex
	pending map[string]time.Time
}

func NewMemoryStateService(secret string, ttl time.Duration) *MemoryStateService {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &MemoryStateService{
		secret:  []byte(secret),
		ttl:     ttl,
		now:     time.Now,
		pending: make(map[string]time.Time),
	}
}

func (s *MemoryStateService) Generate(input OAuthStatePayload) (string, error) {
	if len(s.secret) == 0 {
		return "", ErrOAuthNotConfigured
	}

	nonceBytes := make([]byte, 24)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	expiresAt := s.now().Add(s.ttl)

	env := stateEnvelope{
		Nonce:     nonce,
		UserID:    input.UserID.String(),
		ExpiresAt: expiresAt.Unix(),
	}
	if input.OrganizationID != nil {
		orgID := input.OrganizationID.String()
		env.OrganizationID = &orgID
	}

	payloadJSON, err := json.Marshal(env)
	if err != nil {
		return "", err
	}

	signature := s.sign(payloadJSON)
	state := base64.RawURLEncoding.EncodeToString(payloadJSON) + "." + base64.RawURLEncoding.EncodeToString(signature)

	s.mu.Lock()
	s.pending[nonce] = expiresAt
	s.cleanupExpiredLocked()
	s.mu.Unlock()

	return state, nil
}

func (s *MemoryStateService) Consume(state string) (OAuthStatePayload, error) {
	if len(s.secret) == 0 {
		return OAuthStatePayload{}, ErrOAuthNotConfigured
	}

	parts := splitState(state)
	if len(parts) != 2 {
		return OAuthStatePayload{}, ErrInvalidState
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return OAuthStatePayload{}, ErrInvalidState
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return OAuthStatePayload{}, ErrInvalidState
	}

	if !hmac.Equal(sig, s.sign(payloadJSON)) {
		return OAuthStatePayload{}, ErrInvalidState
	}

	var env stateEnvelope
	if err := json.Unmarshal(payloadJSON, &env); err != nil {
		return OAuthStatePayload{}, ErrInvalidState
	}

	userID, err := uuid.Parse(env.UserID)
	if err != nil {
		return OAuthStatePayload{}, ErrInvalidState
	}

	expiresAt := time.Unix(env.ExpiresAt, 0)
	if s.now().After(expiresAt) {
		return OAuthStatePayload{}, ErrStateExpired
	}

	s.mu.Lock()
	expiry, ok := s.pending[env.Nonce]
	if !ok {
		s.mu.Unlock()
		return OAuthStatePayload{}, ErrInvalidState
	}
	if s.now().After(expiry) {
		delete(s.pending, env.Nonce)
		s.mu.Unlock()
		return OAuthStatePayload{}, ErrStateExpired
	}
	delete(s.pending, env.Nonce)
	s.cleanupExpiredLocked()
	s.mu.Unlock()

	payload := OAuthStatePayload{UserID: userID}
	if env.OrganizationID != nil {
		orgID, parseErr := uuid.Parse(*env.OrganizationID)
		if parseErr != nil {
			return OAuthStatePayload{}, ErrInvalidState
		}
		payload.OrganizationID = &orgID
	}
	return payload, nil
}

func (s *MemoryStateService) sign(payload []byte) []byte {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write(payload)
	return mac.Sum(nil)
}

func (s *MemoryStateService) cleanupExpiredLocked() {
	now := s.now()
	for nonce, expiry := range s.pending {
		if now.After(expiry) {
			delete(s.pending, nonce)
		}
	}
}

func splitState(state string) []string {
	for i := 0; i < len(state); i++ {
		if state[i] == '.' {
			return []string{state[:i], state[i+1:]}
		}
	}
	return nil
}

var _ StateService = (*MemoryStateService)(nil)
