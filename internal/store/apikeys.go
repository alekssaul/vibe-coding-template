package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/alekssaul/template/internal/model"
	"github.com/alekssaul/template/internal/store/db"
	"github.com/google/uuid"
)

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func mapAPIKey(ak db.ApiKey) *model.APIKey {
	return &model.APIKey{
		ID:         ak.ID,
		Name:       ak.Name,
		KeyHash:    ak.KeyHash,
		Permission: model.Permission(ak.Permission),
		CreatedAt:  ak.CreatedAt.Time,
	}
}

// CreateAPIKey generates a new UUID key, stores its hash, and returns the plaintext key once.
func (s *Store) CreateAPIKey(ctx context.Context, req *model.CreateAPIKeyRequest) (*model.CreateAPIKeyResponse, error) {
	key := uuid.New().String()
	hash := hashKey(key)

	ak, err := s.queries.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		Name:       req.Name,
		KeyHash:    hash,
		Permission: string(req.Permission),
	})
	if err != nil {
		return nil, err
	}

	return &model.CreateAPIKeyResponse{
		APIKey: *mapAPIKey(ak),
		Key:    key,
	}, nil
}

// ValidateAPIKey looks up a key by its hash and returns the APIKey metadata.
func (s *Store) ValidateAPIKey(ctx context.Context, key string) (*model.APIKey, error) {
	hash := hashKey(key)
	ak, err := s.queries.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		return nil, err // Let handlers wrap it to 401
	}
	return mapAPIKey(ak), nil
}

// ListAPIKeys returns all API keys (hashes omitted by JSON tags on model).
func (s *Store) ListAPIKeys(ctx context.Context) ([]*model.APIKey, error) {
	dbKeys, err := s.queries.ListAPIKeys(ctx)
	if err != nil {
		return nil, err
	}

	var keys []*model.APIKey
	for _, ak := range dbKeys {
		keys = append(keys, mapAPIKey(ak))
	}
	return keys, nil
}

// DeleteAPIKey removes an API key by ID.
func (s *Store) DeleteAPIKey(ctx context.Context, id int64) error {
	return s.queries.DeleteAPIKey(ctx, id)
}

// CountAPIKeys returns the total number of API keys.
func (s *Store) CountAPIKeys(ctx context.Context) (int, error) {
	count, err := s.queries.CountAPIKeys(ctx)
	return int(count), err
}
