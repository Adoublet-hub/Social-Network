package controllers

import (
	"backend/pkg/models"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
)

func (s *MyServer) CreateCommentPostsGroup() http.HandlerFunc {
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

		username, err := s.getUsernameByUserID(userID)
		if err != nil {
			http.Error(w, "Failed to get username", http.StatusInternalServerError)
			return
		}

		var comment models.CommentPostGroup
		if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		comment.ID = uuid.Must(uuid.NewV4())
		comment.UserID = userID
		comment.Username = username
		comment.CreatedAt = time.Now()

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		query := `INSERT INTO group_posts_comments (id, post_id, content, user_id, username, created_at) VALUES (?, ?, ?, ?, ?, ?)`
		_, err = DB.Exec(query, comment.ID, comment.PostID, comment.Content, comment.UserID, comment.Username, comment.CreatedAt)
		if err != nil {
			http.Error(w, "Failed to create comment", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"comment": comment,
		})
	}
}

func (s *MyServer) ListCommentsByPostGroupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		postIDStr := r.URL.Query().Get("post_id")
		if postIDStr == "" {
			http.Error(w, "Post ID not provided", http.StatusBadRequest)
			return
		}

		postID, err := uuid.FromString(postIDStr)
		if err != nil {
			http.Error(w, "Invalid Post ID", http.StatusBadRequest)
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

		// ✅ Jointure pour récupérer le nom d'utilisateur
		query := `
			SELECT c.id, c.post_id, c.content, c.user_id, u.username, c.created_at
			FROM group_posts_comments AS c
			INNER JOIN users AS u ON c.user_id = u.id
			WHERE c.post_id = ?
			ORDER BY c.created_at DESC
			LIMIT ? OFFSET ?
		`

		rows, err := DB.Query(query, postID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var comments []models.CommentPostGroup
		for rows.Next() {
			var comment models.CommentPostGroup
			if err := rows.Scan(&comment.ID, &comment.PostID, &comment.Content, &comment.UserID, &comment.Username, &comment.CreatedAt); err != nil {
				http.Error(w, "Failed to scan comment", http.StatusInternalServerError)
				return
			}
			comments = append(comments, comment)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(comments); err != nil {
			http.Error(w, "Failed to encode comments", http.StatusInternalServerError)
		}
	}
}

func (s *MyServer) getUsernameByUserID(userID uuid.UUID) (string, error) {
	DB, err := s.Store.OpenDatabase()
	if err != nil {
		return "", err
	}
	defer DB.Close()

	var username string
	query := `SELECT username FROM users WHERE id = ?`
	err = DB.QueryRow(query, userID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}
