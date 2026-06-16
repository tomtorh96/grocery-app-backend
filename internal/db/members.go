package db

import (
	"context"
	"fmt"

	"github.com/tomtorh96/grocery-app/internal/models"
)

// GetMembers returns all members of a list with their roles and usernames
func GetMembers(ctx context.Context, listID string) ([]*models.Member, error) {
	rows, err := Pool.Query(ctx,
		`SELECT lm.user_id, u.username, lm.role, lm.joined_at
		 FROM list_members lm
		 JOIN users u ON lm.user_id = u.id
		 WHERE lm.list_id = $1
		 ORDER BY lm.joined_at ASC`,
		listID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}
	defer rows.Close()

	var members []*models.Member
	for rows.Next() {
		var m models.Member
		if err := rows.Scan(&m.UserID, &m.Username, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}
	return members, nil
}

// GetMemberRole returns the role of a user in a list
func GetMemberRole(ctx context.Context, listID, userID string) (string, error) {
	var role string
	err := Pool.QueryRow(ctx,
		`SELECT role FROM list_members WHERE list_id = $1 AND user_id = $2`,
		listID, userID,
	).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("member not found: %w", err)
	}
	return role, nil
}

// RemoveMember removes a user from a list
func RemoveMember(ctx context.Context, listID, userID string) error {
	_, err := Pool.Exec(ctx,
		`DELETE FROM list_members WHERE list_id = $1 AND user_id = $2`,
		listID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	return nil
}

// UpdateMemberRole changes the role of a member
func UpdateMemberRole(ctx context.Context, listID, userID, role string) error {
	_, err := Pool.Exec(ctx,
		`UPDATE list_members SET role = $1 WHERE list_id = $2 AND user_id = $3`,
		role, listID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	return nil
}

// RenameList updates the name of a list
func RenameList(ctx context.Context, listID, name string) error {
	_, err := Pool.Exec(ctx,
		`UPDATE lists SET name = $1 WHERE id = $2`,
		name, listID,
	)
	if err != nil {
		return fmt.Errorf("failed to rename list: %w", err)
	}
	return nil
}

// LeaveList removes a user from a list
func LeaveList(ctx context.Context, listID, userID string) error {
	_, err := Pool.Exec(ctx,
		`DELETE FROM list_members WHERE list_id = $1 AND user_id = $2`,
		listID, userID,
	)
	return err
}

// GetAdminCount returns how many admins a list has
func GetAdminCount(ctx context.Context, listID string) (int, error) {
	var count int
	err := Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM list_members WHERE list_id = $1 AND role = 'admin'`,
		listID,
	).Scan(&count)
	return count, err
}

// PromoteOldestGuest promotes the earliest joined guest to admin
func PromoteOldestGuest(ctx context.Context, listID string) error {
	_, err := Pool.Exec(ctx,
		`UPDATE list_members SET role = 'admin'
         WHERE list_id = $1 AND user_id = (
             SELECT user_id FROM list_members
             WHERE list_id = $1 AND role = 'guest'
             ORDER BY joined_at ASC
             LIMIT 1
         )`,
		listID,
	)
	return err
}
