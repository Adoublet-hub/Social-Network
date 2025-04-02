package controllers

import (
	"backend/pkg/zwt"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponses struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message"`
}

func (s MyServer) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			SendJSONErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var loginData struct {
			Identifier string `json:"email"`
			Password   string `json:"password"`
		}

		err := json.NewDecoder(r.Body).Decode(&loginData)
		if err != nil {
			log.Println("Failed to decode JSON body:", err)
			SendJSONErrorResponse(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		identifier := strings.TrimSpace(loginData.Identifier)
		password := strings.TrimSpace(loginData.Password)

		if identifier == "" || password == "" {
			log.Println("Identifier or password is empty")
			SendJSONErrorResponse(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			SendJSONErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		userID, storedPassword, username, err := getUserCredentials(DB, identifier)
		if err != nil {
			log.Println("User credential error:", err)
			SendJSONErrorResponse(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		if err != nil {
			log.Println("Incorrect password")
			SendJSONErrorResponse(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		}

		token, err := zwt.GenerateJWT(userID, username)
		if err != nil {
			log.Println("Failed to generate token:", err)
			SendJSONErrorResponse(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Printf("User logged in successfully, userID: %s", userID)

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
		})

		http.SetCookie(w, &http.Cookie{
			Name:    "username",
			Value:   username,
			Expires: time.Now().Add(30 * time.Minute),
		})

		SendJSONResponse(w, LoginResponses{Token: token, Message: "Login successful"}, http.StatusOK)
	}
}

// getUserCredentials vérifie l'existence de l'utilisateur et renvoie une erreur générique si l'utilisateur n'existe pas ou si une erreur survient
func getUserCredentials(DB *sql.DB, identifier string) (uuid.UUID, string, string, error) {
	var userID uuid.UUID
	var storedPassword, username string
	var query string

	// si l'identifiant est un email ou un nom d'utilisateur
	if strings.Contains(identifier, "@") {
		query = "SELECT id, password_hash, username FROM users WHERE email = ?"
	} else {
		query = "SELECT id, password_hash, username FROM users WHERE username = ?"
	}

	err := DB.QueryRow(query, identifier).Scan(&userID, &storedPassword, &username)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, "", "", fmt.Errorf("incorrect  or password")
		}
		// masque les erreurs SQL internes
		log.Printf("Database error: %v", err)
		return uuid.Nil, "", "", fmt.Errorf("incorrect username or password")
	}

	return userID, storedPassword, username, nil
}

func SendJSONResponse(w http.ResponseWriter, response LoginResponses, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		SendJSONErrorResponse(w, "Internal server error", http.StatusInternalServerError)
	}
}

func SendJSONErrorResponse(w http.ResponseWriter, message string, statusCode int) {

	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
	}
	err := json.NewEncoder(w).Encode(map[string]string{"error": message})
	if err != nil {
		log.Printf("Failed to encode JSON error response: %v", err)
	}

}

// gère les requêtes de déconnexion
func (s *MyServer) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Supprime le cookie de token
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    "",
			Expires:  time.Unix(0, 0),
			Secure:   false,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})

		SendJSONResponse(w, LoginResponses{Message: "Logged out successfully"}, http.StatusOK)
	}
}
