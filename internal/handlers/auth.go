package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/tomtorh96/grocery-app/internal/auth"
	"github.com/tomtorh96/grocery-app/internal/db"
)

type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, "password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, err := db.CreateUser(r.Context(), req.Username, hash)
	if err != nil {
		http.Error(w, "username already taken", http.StatusConflict)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
	})
}
