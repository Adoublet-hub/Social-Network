package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
)

func (s *MyServer) SearchUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		log.Println("userID :", userID)

		query := r.URL.Query().Get("query")
		if query == "" {
			http.Error(w, `{"error": "Query parameter is required"}`, http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Printf("Database connection error: %v", err)
			http.Error(w, `{"error": "Failed to open database"}`, http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		searchQuery := `SELECT 
            u.id, 
            u.username, 
            u.avatar, 
            CASE 
                WHEN fr.sender_id = ? THEN 1 
                ELSE 0 
            END AS is_request_pending
        FROM users u
        LEFT JOIN follow_requests fr ON fr.receiver_id = u.id AND fr.sender_id = ?
        WHERE u.id != ? AND u.username LIKE ?
        LIMIT ? OFFSET ?;`

		rows, err := DB.Query(searchQuery, userID, userID, userID, query+"%", 10, 0)
		if err != nil {
			log.Printf("SQL query error: %v", err)
			http.Error(w, `{"error": "Failed to search users"}`, http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		users := []struct {
			ID               string         `json:"id"`
			Username         string         `json:"username"`
			Avatar           sql.NullString `json:"avatar"`
			IsRequestPending bool           `json:"is_request_pending"`
		}{}

		for rows.Next() {
			var user struct {
				ID               string         `json:"id"`
				Username         string         `json:"username"`
				Avatar           sql.NullString `json:"avatar"`
				IsRequestPending bool           `json:"is_request_pending"`
			}
			err := rows.Scan(&user.ID, &user.Username, &user.Avatar, &user.IsRequestPending)
			if err != nil {
				log.Printf("Error scanning user: %v", err)
				http.Error(w, `{"error": "Failed to scan user"}`, http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}

		result := make([]struct {
			ID               string `json:"id"`
			Username         string `json:"username"`
			Avatar           string `json:"avatar"`
			IsRequestPending bool   `json:"is_request_pending"`
		}, len(users))

		for i, user := range users {
			result[i] = struct {
				ID               string `json:"id"`
				Username         string `json:"username"`
				Avatar           string `json:"avatar"`
				IsRequestPending bool   `json:"is_request_pending"`
			}{
				ID:               user.ID,
				Username:         user.Username,
				Avatar:           user.Avatar.String,
				IsRequestPending: user.IsRequestPending,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}
