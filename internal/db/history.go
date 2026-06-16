package db

import (
	"context"
	"fmt"

	"github.com/tomtorh96/grocery-app/internal/models"
)

// AddHistory inserts a history entry
func AddHistory(ctx context.Context, listID, userID, action, itemName string) error {
	_, err := Pool.Exec(ctx,
		`INSERT INTO history (list_id, user_id, action, item_name)
		 VALUES ($1, $2, $3, $4)`,
		listID, userID, action, itemName,
	)
	if err != nil {
		return fmt.Errorf("failed to add history: %w", err)
	}
	return nil
}

// GetHistory returns the activity log for a list, most recent first
func GetHistory(ctx context.Context, listID string) ([]*models.History, error) {
	rows, err := Pool.Query(ctx,
		`SELECT h.id, h.list_id, h.user_id, u.username, h.action, h.item_name, h.timestamp
		 FROM history h
		 JOIN users u ON h.user_id = u.id
		 WHERE h.list_id = $1
		 ORDER BY h.timestamp DESC
		 LIMIT 100`,
		listID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}
	defer rows.Close()

	var entries []*models.History
	for rows.Next() {
		var h models.History
		if err := rows.Scan(&h.ID, &h.ListID, &h.UserID, &h.Username, &h.Action, &h.ItemName, &h.Timestamp); err != nil {
			return nil, err
		}
		entries = append(entries, &h)
	}
	return entries, nil
}
