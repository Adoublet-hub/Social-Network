package controllers

import (
	"backend/pkg/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

func (s *MyServer) CreatePostHandlers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse multipart form with a 20MB limit
		err := r.ParseMultipartForm(20 << 20)
		if err != nil {
			log.Println("Failed to parse multipart form:", err)
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		var post models.Post
		post.Title = r.FormValue("title")
		post.Content = r.FormValue("content")
		post.Visibility = r.FormValue("visibility")
		post.Username = r.FormValue("username")

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("User ID not found in context")
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}
		post.UserID = userID
		post.CreatedAt = time.Now()

		if post.Visibility == "almost_private" {
			allowedUsersStr := r.FormValue("allowed_users")
			if allowedUsersStr != "" {
				allowedUsers := strings.Split(allowedUsersStr, ",")
				for _, userIDStr := range allowedUsers {
					allowedUserID, err := uuid.FromString(userIDStr)
					if err != nil {
						log.Println("Invalid allowed user ID:", userIDStr)
						http.Error(w, "Invalid allowed user ID", http.StatusBadRequest)
						return
					}
					post.AllowedUsers = append(post.AllowedUsers, allowedUserID)
				}
			}
		}

		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			if !IsValidImageExtension(handler.Filename) {
				log.Println("Invalid image file extension")
				http.Error(w, "Invalid image file extension", http.StatusBadRequest)
				return
			}

			log.Println("Tentative de création d'une image...")
			imagesPath, err := UploadImages(w, r, "./image_path/")
			if err != nil {
				log.Printf("Erreur lors du téléversement de l'image : %v\n", err)
				http.Error(w, "Failed to upload image", http.StatusInternalServerError)
				return
			}
			log.Println("Image téléversée avec succès :", imagesPath)

			post.ImagePath = imagesPath
		}

		postID, err := s.StorePost(post)
		if err != nil {
			log.Println("Failed to save post:", err)
			http.Error(w, "Failed to save post", http.StatusInternalServerError)
			return
		}
		post.ID = postID

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(post)
	}
}

func (s *MyServer) ListPostHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("User ID not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println("User ID found:", userID)

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database for ListPost:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		tx, err := DB.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		commitErr := func() error {
			if err != nil {
				return tx.Rollback()
			}
			return tx.Commit()
		}

		defer func() {
			if commitErr() != nil {
				log.Println("Transaction failed to commit/rollback")
			}
		}()

		log.Println("Database and table ready")

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

		log.Printf("Fetching visible posts from database (page: %d, limit: %d)\n", page, limit)

		posts, err := GetVisiblePostsWithPagination(DB, userID, limit, offset)
		if err != nil {
			log.Println("Failed to retrieve posts:", err)
			http.Error(w, "Failed to retrieve posts from the database", http.StatusInternalServerError)
			return
		}

		fmt.Println("posts", posts)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(posts)
		if err != nil {
			log.Println("Failed to encode posts to JSON:", err)
			http.Error(w, "Failed to encode posts to JSON", http.StatusInternalServerError)
		}

	}
}

func GetAvatar(db *sql.DB, userID uuid.UUID) (sql.NullString, error) {
	var avatar sql.NullString
	query := `SELECT avatar FROM users WHERE id = ?`
	err := db.QueryRow(query, userID).Scan(&avatar)
	if err != nil {
		return avatar, fmt.Errorf("failed to query avatar: %w", err)
	}
	return avatar, nil
}

/*--------------------------------------------------------------------------------------------------------------------------*/
