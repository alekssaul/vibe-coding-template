package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/alekssaul/template/internal/model"
)

// ListItems returns paginated items and the total count.
func (s *Store) ListItems(ctx context.Context, limit, offset int) ([]*model.Item, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM items").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, description, created_at, updated_at FROM items ORDER BY id DESC LIMIT ? OFFSET ?",
		limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.Item
	for rows.Next() {
		item := &model.Item{}
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// GetItem returns a single item by ID.
func (s *Store) GetItem(ctx context.Context, id int64) (*model.Item, error) {
	item := &model.Item{}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, description, created_at, updated_at FROM items WHERE id = ?", id).
		Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

// CreateItem inserts a new item and returns it.
func (s *Store) CreateItem(ctx context.Context, req *model.CreateItemRequest) (*model.Item, error) {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO items (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)",
		req.Name, req.Description, now, now)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return s.GetItem(ctx, id)
}

// UpdateItem updates name/description of an existing item and returns it.
func (s *Store) UpdateItem(ctx context.Context, id int64, req *model.UpdateItemRequest) (*model.Item, error) {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		"UPDATE items SET name = ?, description = ?, updated_at = ? WHERE id = ?",
		req.Name, req.Description, now, id)
	if err != nil {
		return nil, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	return s.GetItem(ctx, id)
}

// DeleteItem removes an item by ID.
func (s *Store) DeleteItem(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM items WHERE id = ?", id)
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
