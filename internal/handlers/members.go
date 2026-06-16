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

type updateRoleRequest struct {
	Role string `json:"role"`
}

type renameListRequest struct {
	Name string `json:"name"`
}

func GetMembers(w http.ResponseWriter, r *http.Request) {
	listID := chi.URLParam(r, "id")
	userID := auth.GetUserID(r)

	member, err := db.IsMember(r.Context(), listID, userID)
	if err != nil || !member {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	members, err := db.GetMembers(r.Context(), listID)
	if err != nil {
		http.Error(w, "failed to get members", http.StatusInternalServerError)
		return
	}

	if members == nil {
		members = []*models.Member{}
	}

	writeJSON(w, http.StatusOK, members)
}

func RemoveMember(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		targetUserID := chi.URLParam(r, "userId")
		userID := auth.GetUserID(r)

		// check caller is admin
		role, err := db.GetMemberRole(r.Context(), listID, userID)
		if err != nil || role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// get target username for broadcast
		target, err := db.GetUserByID(r.Context(), targetUserID)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		if err := db.RemoveMember(r.Context(), listID, targetUserID); err != nil {
			http.Error(w, "failed to remove member", http.StatusInternalServerError)
			return
		}

		hub.Broadcast(listID, ws.Message{
			Type: "member_removed",
			Payload: map[string]any{
				"user_id":  targetUserID,
				"username": target.Username,
			},
		}, nil)

		w.WriteHeader(http.StatusOK)
	}
}

func UpdateMemberRole(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		targetUserID := chi.URLParam(r, "userId")
		userID := auth.GetUserID(r)

		// check caller is admin
		role, err := db.GetMemberRole(r.Context(), listID, userID)
		if err != nil || role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		var req updateRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Role != "admin" && req.Role != "guest" {
			http.Error(w, "role must be admin or guest", http.StatusBadRequest)
			return
		}

		if err := db.UpdateMemberRole(r.Context(), listID, targetUserID, req.Role); err != nil {
			http.Error(w, "failed to update role", http.StatusInternalServerError)
			return
		}

		target, _ := db.GetUserByID(r.Context(), targetUserID)

		hub.Broadcast(listID, ws.Message{
			Type: "member_role_changed",
			Payload: map[string]any{
				"user_id":  targetUserID,
				"username": target.Username,
				"role":     req.Role,
			},
		}, nil)

		w.WriteHeader(http.StatusOK)
	}
}

func RenameList(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		userID := auth.GetUserID(r)

		// check caller is admin
		role, err := db.GetMemberRole(r.Context(), listID, userID)
		if err != nil || role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		var req renameListRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name required", http.StatusBadRequest)
			return
		}

		if err := db.RenameList(r.Context(), listID, req.Name); err != nil {
			http.Error(w, "failed to rename list", http.StatusInternalServerError)
			return
		}

		hub.Broadcast(listID, ws.Message{
			Type: "list_renamed",
			Payload: map[string]any{
				"name": req.Name,
			},
		}, nil)

		writeJSON(w, http.StatusOK, map[string]string{"name": req.Name})
	}
}

func LeaveList(hub *ws.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		listID := chi.URLParam(r, "id")
		userID := auth.GetUserID(r)
		username := auth.GetUsername(r)

		role, err := db.GetMemberRole(r.Context(), listID, userID)
		if err != nil {
			http.Error(w, "not a member", http.StatusForbidden)
			return
		}

		if err := db.RemoveMember(r.Context(), listID, userID); err != nil {
			http.Error(w, "failed to leave list", http.StatusInternalServerError)
			return
		}

		// check if any members remain
		members, _ := db.GetMembers(r.Context(), listID)
		if len(members) == 0 {
			// delete the list entirely
			db.DeleteList(r.Context(), listID)
			w.WriteHeader(http.StatusOK)
			return
		}

		// if user was admin and no admins left, promote oldest guest
		if role == "admin" {
			count, _ := db.GetAdminCount(r.Context(), listID)
			if count == 0 {
				db.PromoteOldestGuest(r.Context(), listID)
				hub.Broadcast(listID, ws.Message{
					Type:    "member_role_changed",
					Payload: map[string]any{"promoted": true},
				}, nil)
			}
		}

		hub.Broadcast(listID, ws.Message{
			Type: "member_removed",
			Payload: map[string]any{
				"user_id":  userID,
				"username": username,
			},
		}, nil)

		w.WriteHeader(http.StatusOK)
	}
}
