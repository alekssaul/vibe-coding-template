package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"

	"github.com/alekssaul/template/internal/model"
	"github.com/google/uuid"
)

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// CreateAPIKey generates a new UUID key, stores its hash, and returns the plaintext key once.
func (s *Store) CreateAPIKey(ctx context.Context, req *model.CreateAPIKeyRequest) (*model.CreateAPIKeyResponse, error) {
	key := uuid.New().String()
	hash := hashKey(key)

	result, err := s.db.ExecContext(ctx,
		"INSERT INTO api_keys (name, key_hash, permission) VALUES (?, ?, ?)",
		req.Name, hash, string(req.Permission))
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	ak, err := s.getAPIKeyByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.CreateAPIKeyResponse{APIKey: *ak, Key: key}, nil
}

func (s *Store) getAPIKeyByID(ctx context.Context, id int64) (*model.APIKey, error) {
	ak := &model.APIKey{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, key_hash, permission, created_at FROM api_keys WHERE id = ?", id).
		Scan(&ak.ID, &ak.Name, &ak.KeyHash, &ak.Permission, &ak.CreatedAt)
	return ak, err
}

// ValidateAPIKey looks up a key by its hash and returns the APIKey metadata.
func (s *Store) ValidateAPIKey(ctx context.Context, key string) (*model.APIKey, error) {
	hash := hashKey(key)
	ak := &model.APIKey{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, key_hash, permission, created_at FROM api_keys WHERE key_hash = ?", hash).
		Scan(&ak.ID, &ak.Name, &ak.KeyHash, &ak.Permission, &ak.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid api key")
	}
	return ak, err
}

// ListAPIKeys returns all API keys (hashes omitted by JSON tags on model).
func (s *Store) ListAPIKeys(ctx context.Context) ([]*model.APIKey, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, key_hash, permission, created_at FROM api_keys ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*model.APIKey
	for rows.Next() {
		ak := &model.APIKey{}
		if err := rows.Scan(&ak.ID, &ak.Name, &ak.KeyHash, &ak.Permission, &ak.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, ak)
	}
	return keys, rows.Err()
}

// CountAPIKeys returns the total number of API keys.
func (s *Store) CountAPIKeys(ctx context.Context) (int, error) {
	var n int
	return n, s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM api_keys").Scan(&n)
}

// DeleteAPIKey removes an API key by ID.
func (s *Store) DeleteAPIKey(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM api_keys WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
