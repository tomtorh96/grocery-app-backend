package db

import (
	"context"
	"fmt"

	"github.com/tomtorh96/grocery-app/internal/models"
)

// CreateUser inserts a new user into the database
func CreateUser(ctx context.Context, username, passwordHash string) (*models.User, error) {
	var user models.User
	err := Pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash)
		 VALUES ($1, $2)
		 RETURNING id, username, created_at`,
		username, passwordHash,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return &user, nil
}

// GetUserByUsername fetches a user by their username
func GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := Pool.QueryRow(ctx,
		`SELECT id, username, password_hash, created_at
		 FROM users WHERE username = $1`,
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetUserByID fetches a user by their ID
func GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := Pool.QueryRow(ctx,
		`SELECT id, username, created_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}
