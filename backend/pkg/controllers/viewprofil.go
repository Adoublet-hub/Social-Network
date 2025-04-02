package controllers

import (
	"backend/pkg/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gofrs/uuid"
)

func (s *MyServer) GetUserProfilHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifie la méthode HTTP
		if r.Method != http.MethodGet {
			http.Error(w, `{"error": "Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}

		// Utilisation d'une regex pour extraire l'ID
		re := regexp.MustCompile(`^/viewprofil/([a-zA-Z0-9-]+)$`)
		matches := re.FindStringSubmatch(r.URL.Path)
		if len(matches) < 2 {
			http.Error(w, `{"error": "User ID is required"}`, http.StatusBadRequest)
			return
		}
		userID := matches[1]

		log.Printf("Received request to fetch profile for userID: %s", userID)

		if userID == "" || userID == "undefined" {
			http.Error(w, `{"error": "Invalid user ID"}`, http.StatusBadRequest)
			return
		}

		// Connexion à la base de données
		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, `{"error": "Internal server error"}`, http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		tx, err := DB.Begin()
		if err != nil {
			http.Error(w, `{"error": "Failed to start transaction"}`, http.StatusInternalServerError)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		var user models.UserProfil
		query := `SELECT id, username, first_name, last_name, bio, is_private, avatar FROM users WHERE id = ?`
		err = tx.QueryRow(query, userID).Scan(
			&user.UserID, &user.Username, &user.FirstName, &user.LastName, &user.Bio, &user.IsPrivate, &user.Avatar,
		)
		if err != nil {
			http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
			return
		}

		user.Followers, err = GetFollowers(DB, user.UserID)
		if err != nil {
			http.Error(w, `{"error": "Failed to load followers"}`, http.StatusInternalServerError)
			return
		}

		user.Following, err = GetFollowing(DB, user.UserID)
		if err != nil {
			http.Error(w, `{"error": "Failed to load following"}`, http.StatusInternalServerError)
			return
		}

		user.Posts, err = GetUserPosts(DB, user.UserID)
		if err != nil {
			http.Error(w, `{"error": "Failed to load posts"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
		}
	}
}

func GetUserProfilFromDB(db *sql.DB, userID, loggedInUserID uuid.UUID) (models.UserProfil, error) {

	var profil models.UserProfil

	log.Printf("Querying user profile with ID: %s", userID)

	query := `SELECT id, username, first_name, last_name, bio, is_private, avatar FROM users WHERE id = ?`
	err := db.QueryRow(query, userID).Scan(
		&profil.UserID,
		&profil.Username,
		&profil.FirstName,
		&profil.LastName,
		&profil.Bio,
		&profil.IsPrivate,
		&profil.Avatar,
	)
	if err != nil {
		return profil, fmt.Errorf("failed to query user Profil: %w", err)
	}

	if profil.IsPrivate && !IsUserFollower(db, userID, loggedInUserID) {
		return profil, fmt.Errorf("This profile is private")
	}

	profil.Followers, err = GetFollowers(db, userID)
	if err != nil {
		return profil, fmt.Errorf("failed to get followers: %w", err)
	}

	profil.Following, err = GetFollowing(db, userID)
	if err != nil {
		return profil, fmt.Errorf("failed to get following: %w", err)
	}

	profil.Posts, err = GetUserPosts(db, userID)
	if err != nil {
		return profil, fmt.Errorf("failed to get user posts: %w", err)
	}
	return profil, nil
}

func IsUserFollower(db *sql.DB, userID, followerID uuid.UUID) bool {
	var count int
	query := `SELECT COUNT(*) FROM followers WHERE followed_id = ? AND follower_id = ? AND status = 'accepted'`
	err := db.QueryRow(query, userID, followerID).Scan(&count)
	return err == nil && count > 0
}

func GetFollowers(db *sql.DB, userID uuid.UUID) ([]models.SimpleUser, error) {
	var followers []models.SimpleUser
	query := `SELECT u.id, u.username 
			  FROM users u 
			  INNER JOIN followers f ON u.id = f.follower_id 
			  WHERE f.followed_id = ? AND f.status = 'accepted'`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.SimpleUser
		if err := rows.Scan(&user.UserID, &user.Username); err != nil {
			return nil, err
		}
		followers = append(followers, user)
	}
	return followers, nil
}

func GetFollowing(db *sql.DB, userID uuid.UUID) ([]models.SimpleUser, error) {
	var following []models.SimpleUser
	query := `SELECT u.id, u.username 
			  FROM users u 
			  INNER JOIN followers f ON u.id = f.followed_id 
			  WHERE f.follower_id = ? AND f.status = 'accepted'`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.SimpleUser
		if err := rows.Scan(&user.UserID, &user.Username); err != nil {
			return nil, err
		}
		following = append(following, user)
	}
	return following, nil
}

func GetUserPosts(db *sql.DB, userID uuid.UUID) ([]models.Post, error) {
	var posts []models.Post
	query := `SELECT id, title, content, created_at, visibility, image_path FROM posts WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post models.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.Visibility, &post.ImagePath); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}
