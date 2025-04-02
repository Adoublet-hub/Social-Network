package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
)

func (s *MyServer) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println(" User ID not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println("🔍 User ID found:", userID)

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Printf(" Failed to open database: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		page, limit := 1, 10
		queryParams := r.URL.Query()

		if p := queryParams.Get("page"); p != "" {
			if val, err := strconv.Atoi(p); err == nil && val > 0 {
				page = val
			}
		}
		if l := queryParams.Get("limit"); l != "" {
			if val, err := strconv.Atoi(l); err == nil && val > 0 {
				limit = val
			}
		}

		offset := (page - 1) * limit
		log.Printf("📦 Fetching users (page: %d, limit: %d, offset: %d)\n", page, limit, offset)

		query := `
			SELECT 
				u.id, 
				u.username, 
				u.avatar, 
				EXISTS (
					SELECT 1 FROM follow_requests fr 
					WHERE fr.sender_id = ? AND fr.receiver_id = u.id
				) AS is_request_pending,
				EXISTS (
					SELECT 1 FROM followers f 
					WHERE f.follower_id = ? AND f.followed_id = u.id
				) AS is_following
			FROM users u
			WHERE u.id != ?
			LIMIT ? OFFSET ?;
		`

		rows, err := DB.Query(query, userID, userID, userID, limit, offset)
		if err != nil {
			log.Printf(" Error fetching users: %v\n", err)
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []map[string]interface{}
		for rows.Next() {
			var id, username string
			var avatar sql.NullString
			var isRequestPending, isFollowing bool

			if err := rows.Scan(&id, &username, &avatar, &isRequestPending, &isFollowing); err != nil {
				log.Printf(" Error scanning user data: %v\n", err)
				http.Error(w, "Failed to read user data", http.StatusInternalServerError)
				return
			}

			users = append(users, map[string]interface{}{
				"id":               id,
				"username":         username,
				"avatar":           avatar.String,
				"isRequestPending": isRequestPending,
				"isFollowing":      isFollowing,
			})
		}

		response := map[string]interface{}{
			"page":    page,
			"limit":   limit,
			"results": users,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf(" Error encoding response: %v\n", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func (s *MyServer) ListAmis() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Processing ListAmis request")

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("User ID not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		log.Println("User ID found:", userID)

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Printf("Failed to open database: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		page := 1
		limit := 10
		queryParams := r.URL.Query()

		if p := queryParams.Get("page"); p != "" {
			page, err = strconv.Atoi(p)
			if err != nil || page < 1 {
				page = 1
			}
		}
		if l := queryParams.Get("limit"); l != "" {
			limit, err = strconv.Atoi(l)
			if err != nil || limit < 1 {
				limit = 10
			}
		}
		offset := (page - 1) * limit

		log.Printf("Fetching friends (page: %d, limit: %d, offset: %d)\n", page, limit, offset)

		query := `
            SELECT 
                u.id, 
                u.username, 
                u.avatar  
            FROM users u
            INNER JOIN followers f ON f.followed_id = u.id
            WHERE f.follower_id = ?
            LIMIT ? OFFSET ?;
        `

		rows, err := DB.Query(query, userID, limit, offset)
		if err != nil {
			log.Printf("Error fetching friends: %v\n", err)
			http.Error(w, "Failed to fetch friends", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var result []map[string]interface{}
		for rows.Next() {
			var id, username string
			var avatar sql.NullString

			if err := rows.Scan(&id, &username, &avatar); err != nil {
				log.Printf("Error scanning friend data: %v\n", err)
				http.Error(w, "Failed to read friend data", http.StatusInternalServerError)
				return
			}

			result = append(result, map[string]interface{}{
				"id":       id,
				"username": username,
				"avatar":   avatar.String,
			})
		}

		response := map[string]interface{}{
			"page":    page,
			"limit":   limit,
			"results": result,
		}

		log.Println("Response:", response)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding response: %v\n", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
