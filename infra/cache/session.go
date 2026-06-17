package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultTokenTTL = 25 * time.Hour

type SessionStore struct {
	client *redis.Client
}

func NewSessionStore(client *redis.Client) *SessionStore {
	return &SessionStore{client: client}
}

func tokenRevokeKey(accountID string) string {
	return fmt.Sprintf("session:revoked:%s", accountID)
}

func (s *SessionStore) RevokeToken(ctx context.Context, accountID, token string) error {
	return s.client.Set(ctx, tokenRevokeKey(accountID), token, defaultTokenTTL).Err()
}

func (s *SessionStore) IsTokenRevoked(ctx context.Context, accountID, token string) (bool, error) {
	val, err := s.client.Get(ctx, tokenRevokeKey(accountID)).Result()
	if err != nil || val == "" {
		return false, nil
	}
	return val == token, nil
}
