package controllers

import (
	"backend/pkg/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
)

func (s *MyServer) MyProfil() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("User ID not found in context")
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database for MyProfil", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		tx, err := DB.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		commitErr := func() error {
			if err != nil {
				log.Println("Rolling back transaction due to error:", err)
				return tx.Rollback()
			}
			return tx.Commit()
		}

		defer func() {
			if commitErr() != nil {
				log.Println("Transaction failed to commit/rollback")
			}
		}()

		queryParams := r.URL.Query()
		limit, err := strconv.Atoi(queryParams.Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		}

		offset, err := strconv.Atoi(queryParams.Get("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}

		log.Println("Fetching user profile")
		log.Println("User ID:", userID)
		log.Println("Pagination limit:", limit, "offset:", offset)

		profil, err := GetMyProfil(DB, userID, limit, offset)
		if err != nil {
			http.Error(w, "Failed to get MyProfil", http.StatusInternalServerError)
			return
		}

		profilJSON := models.UserProfilJSON{
			UserID:         profil.UserID,
			Username:       profil.Username,
			FirstName:      profil.FirstName,
			LastName:       profil.LastName,
			Bio:            "",
			Avatar:         profil.Avatar,
			IsPrivate:      profil.IsPrivate,
			FollowersCount: len(profil.Followers),
			FollowingCount: len(profil.Following),
			Followers:      profil.Followers,
			Following:      profil.Following,
			Posts:          profil.Posts,
		}

		if profil.Bio.Valid {
			profilJSON.Bio = profil.Bio.String
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(profilJSON); err != nil {
			log.Println("Failed to encode response to JSON:", err)
			http.Error(w, "Failed to encode response as JSON", http.StatusInternalServerError)
		}
	}
}

func GetMyProfil(db *sql.DB, userID uuid.UUID, limit int, offset int) (models.UserProfil, error) {
	var profil models.UserProfil

	query := `SELECT id, username, first_name, last_name, bio, avatar FROM users WHERE id = ?`
	err := db.QueryRow(query, userID).Scan(&profil.UserID, &profil.Username, &profil.FirstName, &profil.LastName, &profil.Bio, &profil.Avatar)
	if err != nil {
		return profil, fmt.Errorf("failed to query user profile: %w", err)
	}

	profil.Followers, err = GetFollowers(db, userID)
	if err != nil {
		return profil, fmt.Errorf("failed to get followers: %w", err)
	}

	profil.Following, err = GetFollowing(db, userID)
	if err != nil {
		return profil, fmt.Errorf("failed to get following: %w", err)
	}

	profil.Posts, err = GetProfilPostsWithPagination(db, userID, limit, offset)
	if err != nil {
		return profil, fmt.Errorf("failed to get user posts: %w", err)
	}

	return profil, nil
}
