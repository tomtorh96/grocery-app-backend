package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tomtorh96/grocery-app/internal/auth"
	"github.com/tomtorh96/grocery-app/internal/db"
	"github.com/tomtorh96/grocery-app/internal/models"
)

type createListRequest struct {
	Name string `json:"name"`
}

type joinListRequest struct {
	InviteCode string `json:"invite_code"`
}

func CreateList(w http.ResponseWriter, r *http.Request) {
	var req createListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "list name required", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserID(r)
	inviteCode := generateInviteCode()

	list, err := db.CreateList(r.Context(), req.Name, userID, inviteCode)
	if err != nil {
		http.Error(w, "failed to create list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, list)
}

func GetLists(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r)

	lists, err := db.GetListsByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get lists", http.StatusInternalServerError)
		return
	}

	if lists == nil {
		lists = []*models.List{}
	}

	writeJSON(w, http.StatusOK, lists)
}

func GetList(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	member, err := db.IsMember(r.Context(), listID, userID)
	if err != nil || !member {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	list, err := db.GetListByID(r.Context(), listID)
	if err != nil {
		http.Error(w, "list not found", http.StatusNotFound)
		return
	}

	items, err := db.GetItemsByList(r.Context(), listID)
	if err != nil {
		http.Error(w, "failed to get items", http.StatusInternalServerError)
		return
	}

	if items == nil {
		items = []*models.Item{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"list":  list,
		"items": items,
	})
}

func JoinList(w http.ResponseWriter, r *http.Request) {
	var req joinListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserID(r)

	list, err := db.GetListByInviteCode(r.Context(), req.InviteCode)
	if err != nil {
		http.Error(w, "invalid invite code", http.StatusNotFound)
		return
	}

	if err := db.JoinList(r.Context(), list.ID, userID); err != nil {
		http.Error(w, "failed to join list", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, list)
}

// generateInviteCode creates a random 8 character alphanumeric code
func generateInviteCode() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 8)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
