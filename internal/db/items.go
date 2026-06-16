package db

import (
	"context"
	"fmt"

	"github.com/tomtorh96/grocery-app/internal/models"
)

// AddItem inserts a new item into a list
func AddItem(ctx context.Context, listID, name, addedBy, label string) (*models.Item, error) {
	var item models.Item
	err := Pool.QueryRow(ctx,
		`INSERT INTO items (list_id, name, added_by, label)
         VALUES ($1, $2, $3, $4)
         RETURNING id, list_id, name, added_by, is_got, label, created_at`,
		listID, name, addedBy, label,
	).Scan(&item.ID, &item.ListID, &item.Name, &item.AddedBy, &item.IsGot, &item.Label, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to add item: %w", err)
	}
	return &item, nil
}

// GetItemsByList returns all items for a list
func GetItemsByList(ctx context.Context, listID string) ([]*models.Item, error) {
	rows, err := Pool.Query(ctx,
		`SELECT id, list_id, name, added_by, is_got, label, created_at
         FROM items WHERE list_id = $1
         ORDER BY created_at ASC`,
		listID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.ListID, &item.Name, &item.AddedBy, &item.IsGot, &item.Label, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

// GetItemByID returns a single item by ID
func GetItemByID(ctx context.Context, itemID string) (*models.Item, error) {
	var item models.Item
	err := Pool.QueryRow(ctx,
		`SELECT id, list_id, name, added_by, is_got, label, created_at
         FROM items WHERE id = $1`,
		itemID,
	).Scan(&item.ID, &item.ListID, &item.Name, &item.AddedBy, &item.IsGot, &item.Label, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}
	return &item, nil
}

// DeleteItem removes an item from a list
func DeleteItem(ctx context.Context, itemID string) error {
	_, err := Pool.Exec(ctx,
		`DELETE FROM items WHERE id = $1`,
		itemID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

// MarkItem sets the is_got status of an item
func MarkItem(ctx context.Context, itemID string, isGot bool) (*models.Item, error) {
	var item models.Item
	err := Pool.QueryRow(ctx,
		`UPDATE items SET is_got = $1
         WHERE id = $2
         RETURNING id, list_id, name, added_by, is_got, label, created_at`,
		isGot, itemID,
	).Scan(&item.ID, &item.ListID, &item.Name, &item.AddedBy, &item.IsGot, &item.Label, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to mark item: %w", err)
	}
	return &item, nil
}

// ResetAllItems sets all items in a list to is_got = false
func ResetAllItems(ctx context.Context, listID string) error {
	_, err := Pool.Exec(ctx,
		`UPDATE items SET is_got = false WHERE list_id = $1`,
		listID,
	)
	if err != nil {
		return fmt.Errorf("failed to reset items: %w", err)
	}
	return nil
}

func RenameItem(ctx context.Context, itemID, name, label string) (*models.Item, error) {
	var item models.Item
	err := Pool.QueryRow(ctx,
		`UPDATE items SET name = $1, label = $2
         WHERE id = $3
         RETURNING id, list_id, name, added_by, is_got, label, created_at`,
		name, label, itemID,
	).Scan(&item.ID, &item.ListID, &item.Name, &item.AddedBy, &item.IsGot, &item.Label, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to rename item: %w", err)
	}
	return &item, nil
}
