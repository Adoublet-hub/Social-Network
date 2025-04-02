package controllers

import (
	"backend/pkg/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gofrs/uuid"
)

func (s *MyServer) LikeComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestData models.CommentLike

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			log.Println("Error decoding JSON:", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		commentID, err := uuid.FromString(requestData.CommentID.String())
		if err != nil {
			http.Error(w, `{"error": "Invalid comment ID"}`, http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("User ID not found in context or is invalid")
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		if err := s.ToggleLikeComment(userID, commentID, "like"); err != nil {
			log.Println("Error toggling like:", err)
			http.Error(w, "Failed to like post", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *MyServer) UnlikeComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var requestData models.CommentLike

		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			log.Println("Error decoding JSON:", err)
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}

		commentID, err := uuid.FromString(requestData.CommentID.String())
		if err != nil {
			log.Println("Invalid comment ID", err)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			log.Println("User ID not found in context or is invalid")
			return
		}

		if err := s.ToggleLikeComment(userID, commentID, "unlike"); err != nil {
			log.Println("Failed to toggle like:", err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *MyServer) ToggleLikeComment(userID, commentID uuid.UUID, interactionType string) error {
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
		if err := ToggleLikeComment(userID, commentID, tx); err != nil {
			tx.Rollback()
			log.Println("Failed to toggle like:", err)
			return err
		}
	case "unlike":
		if err := ToggleUnlikeComment(userID, commentID, tx); err != nil {
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

func ToggleLikeComment(userID, commentID uuid.UUID, tx *sql.Tx) error {
	liked, err := UserLikedComment(userID, commentID, tx)
	if err != nil {
		return err
	}

	if liked {
		return DeleteLikeComment(userID, commentID, tx)
	}

	return CreateLikeComment(userID, commentID, tx)
}

func ToggleUnlikeComment(userID, commentID uuid.UUID, tx *sql.Tx) error {
	liked, err := UserLikedComment(userID, commentID, tx)
	if err != nil {
		return err
	}

	unliked, err := UserUnlikedComment(userID, commentID, tx)
	if err != nil {
		return err
	}

	if liked {
		return DeleteLikeComment(userID, commentID, tx)
	}

	if unliked {
		return DeleteUnlikeComment(userID, commentID, tx)
	}

	return CreateUnlikeComment(userID, commentID, tx)
}

/*--------------------------------------------------------------------*/

func CreateLikeComment(userID, commentID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO comment_interactions (user_id, comment_id, interaction_type) VALUES (?, ?, 'like')", userID, commentID)
	if err != nil {
		return err
	}
	return UpdateCommentLikeCount(tx, commentID, "likes", true)
}

func DeleteLikeComment(userID, commentID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec("DELETE FROM comment_interactions WHERE user_id = ? AND comment_id = ? AND interaction_type = 'like'", userID, commentID)
	if err != nil {
		return err
	}
	return UpdateCommentLikeCount(tx, commentID, "likes", false)
}

func CreateUnlikeComment(userID, commentID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec("INSERT INTO comment_interactions (user_id, comment_id, interaction_type) VALUES (?, ?, 'unlike')", userID, commentID)
	if err != nil {
		return err
	}
	return UpdateCommentLikeCount(tx, commentID, "unlikes", true)
}

func DeleteUnlikeComment(userID, commentID uuid.UUID, tx *sql.Tx) error {
	_, err := tx.Exec("DELETE FROM comment_interactions WHERE user_id = ? AND comment_id = ? AND interaction_type = 'unlike'", userID, commentID)
	if err != nil {
		return err
	}
	return UpdateCommentLikeCount(tx, commentID, "unlikes", false)
}

func UpdateCommentLikeCount(tx *sql.Tx, commentID uuid.UUID, likeType string, increment bool) error {
	operation := "+"
	if !increment {
		operation = "-"
	}

	query := fmt.Sprintf("UPDATE comments SET total_%s = total_%s %s 1 WHERE id = ?", likeType, likeType, operation)
	_, err := tx.Exec(query, commentID)
	return err
}

/*--------------------------------------------------------------------*/

func UserLikedComment(userID, commentID uuid.UUID, tx *sql.Tx) (bool, error) {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM comment_interactions WHERE user_id = ? AND comment_id = ? AND interaction_type = 'like'", userID, commentID).Scan(&count)
	return count > 0, err
}

func UserUnlikedComment(userID, commentID uuid.UUID, tx *sql.Tx) (bool, error) {
	var count int
	err := tx.QueryRow("SELECT COUNT(*) FROM comment_interactions WHERE user_id = ? AND comment_id = ? AND interaction_type = 'unlike'", userID, commentID).Scan(&count)
	return count > 0, err
}
