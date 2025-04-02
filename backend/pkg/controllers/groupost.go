package controllers

import (
	"backend/pkg/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
)

func (s *MyServer) CreatePostGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "User not logged in", http.StatusUnauthorized)
			return
		}

		var postGroup models.PostGroup
		if err := json.NewDecoder(r.Body).Decode(&postGroup); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		username, err := s.getUsernameByUserID(userID)
		if err != nil {
			http.Error(w, "Failed to get username", http.StatusInternalServerError)
			return
		}
		postGroup.Username = username

		postGroup.ID = uuid.Must(uuid.NewV4())
		postGroup.UserID = userID
		postGroup.CreatedAt = time.Now()
		postGroup.UpdatedAt = time.Now()

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		query := `INSERT INTO group_posts (id, group_id, user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
		_, err = DB.Exec(query, postGroup.ID, postGroup.GroupID, postGroup.UserID, postGroup.Title, postGroup.Content, postGroup.CreatedAt, postGroup.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Post created successfully"))
	}
}

func (s *MyServer) ListPostGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		groupIDStr := r.URL.Query().Get("group_id")
		if groupIDStr == "" {
			http.Error(w, "Group ID not provided", http.StatusBadRequest)
			return
		}

		groupID, err := uuid.FromString(groupIDStr)
		if err != nil {
			http.Error(w, "Invalid Group ID", http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
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
		log.Printf("Fetching posts from database (page: %d, limit: %d)\n", page, limit)

		query := `
		SELECT gp.id, gp.group_id, gp.user_id, u.username, gp.title, gp.content, gp.created_at, gp.updated_at
		FROM group_posts gp
		JOIN users u ON gp.user_id = u.id
		WHERE gp.group_id = ?
		LIMIT ? OFFSET ?
	`
		rows, err := DB.Query(query, groupID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var postsGroup []models.PostGroup
		for rows.Next() {
			var postgroup models.PostGroup
			if err := rows.Scan(&postgroup.ID, &postgroup.GroupID, &postgroup.UserID, &postgroup.Username, &postgroup.Title, &postgroup.Content, &postgroup.CreatedAt, &postgroup.UpdatedAt); err != nil {
				http.Error(w, "Failed to scan post", http.StatusInternalServerError)
				return
			}
			postsGroup = append(postsGroup, postgroup)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(postsGroup); err != nil {
			http.Error(w, "Failed to encode posts", http.StatusInternalServerError)
		}
	}
}
