package controllers

import (
	"backend/pkg/models"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
)

func (s *MyServer) UpdateProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, ok := r.Context().Value(userIDKey).(uuid.UUID)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var updatedProfile models.UserProfil
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(body, &updatedProfile); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		query := "UPDATE users SET "
		var params []interface{}
		var updates []string

		// Vérifier si chaque champ est valide avant de l'ajouter
		if updatedProfile.FirstName.Valid {
			updates = append(updates, "first_name = ?")
			params = append(params, updatedProfile.FirstName.String)
		}
		if updatedProfile.LastName.Valid {
			updates = append(updates, "last_name = ?")
			params = append(params, updatedProfile.LastName.String)
		}
		if updatedProfile.Email != "" {
			updates = append(updates, "email = ?")
			params = append(params, updatedProfile.Email)
		}
		if updatedProfile.Gender != "" {
			updates = append(updates, "gender = ?")
			params = append(params, updatedProfile.Gender)
		}
		if updatedProfile.Avatar.Valid {
			updates = append(updates, "avatar = ?")
			params = append(params, updatedProfile.Avatar.String)
		}
		if updatedProfile.Bio.Valid {
			updates = append(updates, "bio = ?")
			params = append(params, updatedProfile.Bio.String)
		}
		if updatedProfile.PhoneNumber.Valid {
			updates = append(updates, "phone_number = ?")
			params = append(params, updatedProfile.PhoneNumber.String)
		}
		if updatedProfile.Address.Valid {
			updates = append(updates, "address = ?")
			params = append(params, updatedProfile.Address.String)
		}

		if len(updates) == 0 {
			http.Error(w, "No fields to update", http.StatusBadRequest)
			return
		}

		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
		query += strings.Join(updates, ", ") + " WHERE id = ?"
		params = append(params, userID)

		_, err = DB.Exec(query, params...)
		if err != nil {
			http.Error(w, "Failed to update profile", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
	}
}

// func checkUser(db *sql.DB, updatedProfile models.UserProfil, userID uuid.UUID) error {
// 	var countEmail, countUsername, countPhone int

// 	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?", updatedProfile.Email, userID).Scan(&countEmail)
// 	if err != nil {
// 		log.Println("Failed to check email existence:", err)
// 		return fmt.Errorf("failed to check email existence: %w", err)
// 	}
// 	if countEmail > 0 {
// 		return fmt.Errorf("email already exists")
// 	}

// 	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? AND id != ?", updatedProfile.Username, userID).Scan(&countUsername)
// 	if err != nil {
// 		log.Println("Failed to check username existence:", err)
// 		return fmt.Errorf("failed to check username existence: %w", err)
// 	}
// 	if countUsername > 0 {
// 		return fmt.Errorf("username already exists")
// 	}

// 	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE phone_number = ? AND id != ?", updatedProfile.PhoneNumber, userID).Scan(&countPhone)
// 	if err != nil {
// 		log.Println("Failed to check phone existence:", err)
// 		return fmt.Errorf("failed to check phone existence: %w", err)
// 	}
// 	if countPhone > 0 {
// 		return fmt.Errorf("phone number already exists")
// 	}
// 	return nil
// }

// func mapToUserProfilResponse(profile models.UserProfil) models.UserProfilResponse {
// 	return models.UserProfilResponse{
// 		UserID:      profile.UserID,
// 		Username:    profile.Username,
// 		FirstName:   getStringValue(profile.FirstName),
// 		LastName:    getStringValue(profile.LastName),
// 		Email:       profile.Email,
// 		Gender:      profile.Gender,
// 		Bio:         getStringValue(profile.Bio),
// 		IsPrivate:   profile.IsPrivate,
// 		Avatar:      getStringValue(profile.Avatar),
// 		PhoneNumber: getStringValue(profile.PhoneNumber),
// 		Address:     getStringValue(profile.Address),
// 		UpdatedAt:   profile.UpdatedAt,
// 	}
// }

// func getStringValue(ns models.NullString) string {
// 	if ns.Valid {
// 		return ns.String
// 	}
// 	return ""
// }
