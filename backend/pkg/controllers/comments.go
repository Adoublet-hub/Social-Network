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

func (s *MyServer) CreateCommentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var comment models.Comment

			if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
				log.Printf("Failed to decode comment request payload: %v", err)
				http.Error(w, "Invalid request comment payload", http.StatusBadRequest)
				return
			}

			// Générer un nouvel ID unique pour le commentaire
			newID, err := uuid.NewV4()
			if err != nil {
				log.Println("Failed to generate new UUID for comment:", err)
				http.Error(w, "Failed to generate unique ID", http.StatusInternalServerError)
				return
			}
			comment.ID = newID

			comment.CreatedAt = time.Now()
			userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
			if !ok || userID == uuid.Nil {
				log.Println("User ID not found in context")
				http.Error(w, "User ID not found in context", http.StatusUnauthorized)
				return
			}

			comment.UserID = userID

			if err := s.StoreComment(comment); err != nil {
				log.Println("Failed to store comment:", err)
				http.Error(w, "Failed to store comment", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(comment)
		} else {
			http.NotFound(w, r)
		}
	}
}

func (s *MyServer) ListCommentHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Récupérer post_id à partir des paramètres de la requête
		postIDStr := r.URL.Query().Get("post_id")
		log.Println("Post ID from query:", postIDStr)
		if postIDStr == "" {
			http.Error(w, "Post ID is required", http.StatusBadRequest)
			return
		}

		postID, err := uuid.FromString(postIDStr)
		if err != nil {
			log.Println("Invalid Post ID format:", err)
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("User ID not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println("User ID:", userID)

		// Connexion à la base de données
		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, "Failed to open database", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		// Pagination
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
		log.Printf("Fetching comments for Post ID: %s (page: %d, limit: %d)\n", postID, page, limit)

		// Récupération des commentaires
		comments, err := GetCommentsByPost(DB, postID, offset, limit, userID)
		if err != nil {
			log.Println("Failed to retrieve comments:", err)
			http.Error(w, "Failed to retrieve comments", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"comments": comments,
			"page":     page,
			"limit":    limit,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println("Failed to encode comments to JSON:", err)
			http.Error(w, "Failed to encode response as JSON", http.StatusInternalServerError)
		}
	}
}
