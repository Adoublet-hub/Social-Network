package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
)

func (s *MyServer) LikePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var requestData struct {
			PostID string `json:"post_id"`
		}

		log.Println("Request Body:", r.Body)

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			log.Println("Error decoding JSON:", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		log.Println("Parsed PostID:", requestData.PostID)

		postID, err := uuid.FromString(requestData.PostID)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}

		if err := s.togglePostLike(userID, postID, "like"); err != nil {
			log.Println("Error toggling like:", err)
			http.Error(w, "Failed to like post", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *MyServer) UnlikePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var requestData struct {
			PostID string `json:"post_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		postID, err := uuid.FromString(requestData.PostID)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("User ID not found in context")
			http.Error(w, "User ID not found in context", http.StatusUnauthorized)
			return
		}

		if err := s.togglePostLike(userID, postID, "unlike"); err != nil {
			log.Println("Error toggling unlike:", err)
			http.Error(w, "Failed to unlike post", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// togglePostLike gère à la fois les "like" et "unlike" en fonction du type d'interaction
func (s *MyServer) togglePostLike(userID, postID uuid.UUID, interactionType string) error {
	DB, err := s.Store.OpenDatabase()
	if err != nil {
		log.Println("Failed to open database:", err)
		return err
	}
	defer DB.Close()

	tx, err := DB.Begin()
	if err != nil {
		log.Println("Failed to begin transaction:", err)
		return err
	}

	switch interactionType {
	case "like":
		if err := ToggleLike(userID, postID, tx); err != nil {
			tx.Rollback()
			log.Println("Failed to toggle like:", err)
			return err
		}
	case "unlike":
		if err := ToggleUnLike(userID, postID, tx); err != nil {
			tx.Rollback()
			log.Println("Failed to toggle unlike:", err)
			return err
		}
	default:
		tx.Rollback()
		return fmt.Errorf("invalid interaction type: %s", interactionType)
	}

	return tx.Commit()
}

/*-------------------------------------------------------------------------*/
// ToggleLike gère l'ajout ou la suppression d'un like sur une publication
func ToggleLike(userID, postID uuid.UUID, tx *sql.Tx) error {
	liked, err := UserLikedPost(userID, postID, tx)
	if err != nil {
		return err
	}

	if liked {
		return DeleteLike(userID, postID, tx)
	}

	return CreateLike(userID, postID, tx)
}

func ToggleUnLike(userID, postID uuid.UUID, tx *sql.Tx) error {
	liked, err := UserLikedPost(userID, postID, tx)
	if err != nil {
		return err
	}

	if liked {
		return DeleteLike(userID, postID, tx)
	}

	return CreateUnLike(userID, postID, tx)
}

/*--------------------------------------------------------------*/

func UpdatePostLikes(tx *sql.Tx, postID uuid.UUID, likeType string, increment bool) error {
	operation := "+"
	if !increment {
		operation = "-"
	}

	query := fmt.Sprintf("UPDATE posts SET total_%s = total_%s %s 1 WHERE id = ?", likeType, likeType, operation)
	_, err := tx.Exec(query, postID)
	return err
}

func CreateLike(userID, postID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO post_interactions (user_id, post_id, interaction_type) VALUES (?, ?, 'like')", userID, postID)
	if err != nil {
		return err
	}
	return UpdatePostLikes(tx, postID, "likes", true)
}

func DeleteLike(userID, postID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec("DELETE FROM post_interactions WHERE user_id = ? AND post_id = ? AND interaction_type = 'like'", userID, postID)
	if err != nil {
		return err
	}
	return UpdatePostLikes(tx, postID, "likes", false)
}

func CreateUnLike(userID, postID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO post_interactions (user_id, post_id, interaction_type) VALUES (?, ?, 'unlike')", userID, postID)
	if err != nil {
		return err
	}
	return UpdatePostLikes(tx, postID, "unlikes", true)
}

func DeleteUnLike(userID, postID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec("DELETE FROM post_interactions WHERE user_id = ? AND post_id = ? AND interaction_type = 'unlike'", userID, postID)
	if err != nil {
		return err
	}
	return UpdatePostLikes(tx, postID, "unlikes", false)
}

// Vérification si l'utilisateur a liké ou unliké un post
func UserLikedPost(userID, postID uuid.UUID, tx *sql.Tx) (bool, error) {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM post_interactions WHERE user_id = ? AND post_id = ? AND interaction_type = 'like'", userID, postID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UserUnLikedPost(userID, postID uuid.UUID, tx *sql.Tx) (bool, error) {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM post_interactions WHERE user_id = ? AND post_id = ? AND interaction_type = 'unlike'", userID, postID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
