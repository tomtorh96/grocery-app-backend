package db

import (
	"context"
	"fmt"

	"github.com/tomtorh96/grocery-app/internal/models"
)

// CreateList inserts a new list and adds the creator as a member
func CreateList(ctx context.Context, name, createdBy, inviteCode string) (*models.List, error) {
	tx, err := Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var list models.List
	err = tx.QueryRow(ctx,
		`INSERT INTO lists (name, invite_code, created_by)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, invite_code, created_by, created_at`,
		name, inviteCode, createdBy,
	).Scan(&list.ID, &list.Name, &list.InviteCode, &list.CreatedBy, &list.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create list: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO list_members (list_id, user_id, role) VALUES ($1, $2, 'admin')`,
		list.ID, createdBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add creator as member: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &list, nil
}

// GetListsByUser returns all lists a user is a member of
func GetListsByUser(ctx context.Context, userID string) ([]*models.List, error) {
	rows, err := Pool.Query(ctx,
		`SELECT l.id, l.name, l.invite_code, l.created_by, l.created_at
		 FROM lists l
		 JOIN list_members lm ON l.id = lm.list_id
		 WHERE lm.user_id = $1
		 ORDER BY l.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get lists: %w", err)
	}
	defer rows.Close()

	var lists []*models.List
	for rows.Next() {
		var list models.List
		if err := rows.Scan(&list.ID, &list.Name, &list.InviteCode, &list.CreatedBy, &list.CreatedAt); err != nil {
			return nil, err
		}
		lists = append(lists, &list)
	}
	return lists, nil
}

// GetListByID returns a single list by ID
func GetListByID(ctx context.Context, listID string) (*models.List, error) {
	var list models.List
	err := Pool.QueryRow(ctx,
		`SELECT id, name, invite_code, created_by, created_at
		 FROM lists WHERE id = $1`,
		listID,
	).Scan(&list.ID, &list.Name, &list.InviteCode, &list.CreatedBy, &list.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("list not found: %w", err)
	}
	return &list, nil
}

// GetListByInviteCode returns a list by its invite code
func GetListByInviteCode(ctx context.Context, inviteCode string) (*models.List, error) {
	var list models.List
	err := Pool.QueryRow(ctx,
		`SELECT id, name, invite_code, created_by, created_at
		 FROM lists WHERE invite_code = $1`,
		inviteCode,
	).Scan(&list.ID, &list.Name, &list.InviteCode, &list.CreatedBy, &list.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("list not found: %w", err)
	}
	return &list, nil
}

// JoinList adds a user as a member of a list
func JoinList(ctx context.Context, listID, userID string) error {
	_, err := Pool.Exec(ctx,
		`INSERT INTO list_members (list_id, user_id)
		 VALUES ($1, $2)
		 ON CONFLICT DO NOTHING`,
		listID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to join list: %w", err)
	}
	return nil
}

// IsMember checks if a user is a member of a list
func IsMember(ctx context.Context, listID, userID string) (bool, error) {
	var count int
	err := Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM list_members
		 WHERE list_id = $1 AND user_id = $2`,
		listID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func DeleteList(ctx context.Context, listID string) error {
	_, err := Pool.Exec(ctx,
		`DELETE FROM lists WHERE id = $1`,
		listID,
	)
	return err
}
