package store

import (
	"context"

	"github.com/alekssaul/template/internal/model"
	"github.com/alekssaul/template/internal/store/db"
)

func mapItem(i db.Item) *model.Item {
	return &model.Item{
		ID:          i.ID,
		Name:        i.Name,
		Description: i.Description,
		CreatedAt:   i.CreatedAt.Time,
		UpdatedAt:   i.UpdatedAt.Time,
	}
}

// ListItems returns paginated items and the total count.
func (s *Store) ListItems(ctx context.Context, limit, offset int) ([]*model.Item, int, error) {
	total, err := s.queries.CountItems(ctx)
	if err != nil {
		return nil, 0, err
	}

	dbItems, err := s.queries.ListItems(ctx, db.ListItemsParams{
		Limit:  int64(limit),
		Offset: int64(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	var items []*model.Item
	for _, i := range dbItems {
		items = append(items, mapItem(i))
	}
	return items, int(total), nil
}

// GetItem returns a single item by ID.
func (s *Store) GetItem(ctx context.Context, id int64) (*model.Item, error) {
	i, err := s.queries.GetItem(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapItem(i), nil
}

// CreateItem inserts a new item and returns it.
func (s *Store) CreateItem(ctx context.Context, req *model.CreateItemRequest) (*model.Item, error) {
	i, err := s.queries.CreateItem(ctx, db.CreateItemParams{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return mapItem(i), nil
}

// UpdateItem updates name/description of an existing item and returns it.
func (s *Store) UpdateItem(ctx context.Context, id int64, req *model.UpdateItemRequest) (*model.Item, error) {
	i, err := s.queries.UpdateItem(ctx, db.UpdateItemParams{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	return mapItem(i), nil
}

// DeleteItem removes an item by ID.
func (s *Store) DeleteItem(ctx context.Context, id int64) error {
	err := s.queries.DeleteItem(ctx, id)
	// sqlc exec doesn't return sql.ErrNoRows if 0 rows affected by default in sqlite,
	// but for simplicity our template can just return err.
	if err != nil {
		return err
	}
	// To strictly emulate the old behavior of returning ErrNoRows:
	// We might need to check if the item existed first or ignore it.
	// For API simplicity, DELETE is idempotent. We'll leave it as is.
	return nil
}
