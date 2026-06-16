package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tomtorh96/grocery-app/internal/auth"
	"github.com/tomtorh96/grocery-app/internal/db"
	"github.com/tomtorh96/grocery-app/internal/models"
	"github.com/tomtorh96/grocery-app/internal/ws"
)

type addItemRequest struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type markItemRequest struct {
	IsGot bool `json:"is_got"`
}

type editItemRequest struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

func AddItem(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		userID := auth.GetUserID(r)
		username := auth.GetUsername(r)

		member, err := db.IsMember(r.Context(), listID, userID)
		if err != nil || !member {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		var req addItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "item name required", http.StatusBadRequest)
			return
		}

		item, err := db.AddItem(r.Context(), listID, req.Name, userID, req.Label)
		if err != nil {
			http.Error(w, "failed to add item", http.StatusInternalServerError)
			return
		}

		db.AddHistory(r.Context(), listID, userID, "added", req.Name)

		hub.Broadcast(listID, ws.Message{
			Type: "item_added",
			Payload: map[string]any{
				"item":     item,
				"added_by": username,
			},
		}, nil)

		writeJSON(w, http.StatusCreated, item)
	}
}

func DeleteItem(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		itemID := chi.URLParam(r, "itemId")
		userID := auth.GetUserID(r)
		username := auth.GetUsername(r)

		member, err := db.IsMember(r.Context(), listID, userID)
		if err != nil || !member {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		item, err := db.GetItemByID(r.Context(), itemID)
		if err != nil {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}

		if err := db.DeleteItem(r.Context(), itemID); err != nil {
			http.Error(w, "failed to delete item", http.StatusInternalServerError)
			return
		}

		db.AddHistory(r.Context(), listID, userID, "removed", item.Name)

		hub.Broadcast(listID, ws.Message{
			Type: "item_removed",
			Payload: map[string]any{
				"item_id":    itemID,
				"removed_by": username,
			},
		}, nil)

		w.WriteHeader(http.StatusOK)
	}
}

func MarkItem(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		itemID := chi.URLParam(r, "itemId")
		userID := auth.GetUserID(r)
		username := auth.GetUsername(r)

		member, err := db.IsMember(r.Context(), listID, userID)
		if err != nil || !member {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		var req markItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		item, err := db.MarkItem(r.Context(), itemID, req.IsGot)
		if err != nil {
			http.Error(w, "failed to mark item", http.StatusInternalServerError)
			return
		}

		action := "marked"
		msgType := "item_marked"
		if !req.IsGot {
			action = "unmarked"
			msgType = "item_unmarked"
		}

		db.AddHistory(r.Context(), listID, userID, action, item.Name)

		hub.Broadcast(listID, ws.Message{
			Type: msgType,
			Payload: map[string]any{
				"item":      item,
				"marked_by": username,
			},
		}, nil)

		writeJSON(w, http.StatusOK, item)
	}
}

func ResetItems(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		userID := auth.GetUserID(r)
		username := auth.GetUsername(r)

		member, err := db.IsMember(r.Context(), listID, userID)
		if err != nil || !member {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if err := db.ResetAllItems(r.Context(), listID); err != nil {
			http.Error(w, "failed to reset items", http.StatusInternalServerError)
			return
		}

		db.AddHistory(r.Context(), listID, userID, "reset_all", "")

		hub.Broadcast(listID, ws.Message{
			Type: "reset_all",
			Payload: map[string]any{
				"reset_by": username,
			},
		}, nil)

		w.WriteHeader(http.StatusOK)
	}
}

func GetHistory(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	member, err := db.IsMember(r.Context(), listID, userID)
	if err != nil || !member {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	history, err := db.GetHistory(r.Context(), listID)
	if err != nil {
		http.Error(w, "failed to get history", http.StatusInternalServerError)
		return
	}

	if history == nil {
		history = []*models.History{}
	}

	writeJSON(w, http.StatusOK, history)
}

func EditItem(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		itemID := chi.URLParam(r, "itemId")
		userID := auth.GetUserID(r)
		username := auth.GetUsername(r)

		member, err := db.IsMember(r.Context(), listID, userID)
		if err != nil || !member {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		var req editItemRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		item, err := db.RenameItem(r.Context(), itemID, req.Name, req.Label)
		if err != nil {
			http.Error(w, "failed to edit item", http.StatusInternalServerError)
			return
		}

		db.AddHistory(r.Context(), listID, userID, "edited", item.Name)

		hub.Broadcast(listID, ws.Message{
			Type: "item_edited",
			Payload: map[string]any{
				"item":      item,
				"edited_by": username,
			},
		}, nil)

		writeJSON(w, http.StatusOK, item)
	}
}
