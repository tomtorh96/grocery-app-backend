package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type List struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	InviteCode string    `json:"invite_code"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type Item struct {
	ID        string    `json:"id"`
	ListID    string    `json:"list_id"`
	Name      string    `json:"name"`
	AddedBy   string    `json:"added_by"`
	IsGot     bool      `json:"is_got"`
	Label     string    `json:"label"`
	CreatedAt time.Time `json:"created_at"`
}

type History struct {
	ID        string    `json:"id"`
	ListID    string    `json:"list_id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	ItemName  string    `json:"item_name"`
	Timestamp time.Time `json:"timestamp"`
}

type ListMember struct {
	ListID   string    `json:"list_id"`
	UserID   string    `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type Member struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

