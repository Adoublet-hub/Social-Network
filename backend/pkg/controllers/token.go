package controllers

import (
	"backend/pkg/zwt"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"
const usernameIDKey contextKey = "username"

func (s *MyServer) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ignore la vérification pour les requêtes OPTIONS (CORS) et la route de déconnexion
		if r.Method == http.MethodOptions || r.URL.Path == "/logout" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("No Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println("Token reçu :", authHeader)

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Println("Invalid Authorization format")
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]
		claims, err := zwt.VerifyJWT(token)
		if err != nil {
			log.Println("Token verification failed:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println("User ID from token:", claims.UserID)
		log.Println("Username from token:", claims.Username)

		// Inject User ID et Username
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, usernameIDKey, claims.Username)
		log.Println("Injecting User ID and Username into context:", claims.UserID, claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))

	}
}

func (s *MyServer) VerifyTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// extraire le token de l'en-tête Authorization
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]
		claims, err := zwt.VerifyJWT(token)
		if err != nil {
			log.Println("Token verification failed:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Si le token est valide, répondre avec un statut 200 ok
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Token is valid. User ID: %s", claims.UserID)))
	}
}

/*
import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Exp      int64     `json:"exp"`
}

var jwtKey = []byte("my_secret_key")

type contextKey string

const userIDKey contextKey = "userID"

// GenerateJWT génère un JWT avec UserID, Username et une date d'expiration
func GenerateJWT(userID uuid.UUID, username string) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	expirationTime := time.Now().Add(24 * time.Hour).Unix()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Exp:      expirationTime,
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signature := GenerateHMACSHA256(header + "." + payload)
	token := fmt.Sprintf("%s.%s.%s", header, payload, signature)

	// Log parts pour verification
	log.Println("Token generation:")
	log.Println("Header:", header)
	log.Println("Payload:", payload)
	log.Println("Signature:", signature)

	return token, nil
}

// generateHMACSHA256 génère la signature HMAC-SHA256
func GenerateHMACSHA256(data string) string {
	h := hmac.New(sha256.New, jwtKey)
	h.Write([]byte(data))
	return strings.TrimRight(base64.URLEncoding.EncodeToString(h.Sum(nil)), "=")
}

// VerifyJWT vérifie le token JWT et retourne les claims si valides
func VerifyJWT(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		log.Println("Invalid token format")
		return nil, fmt.Errorf("invalid token format")
	}

	header := parts[0]
	payload := parts[1]
	signature := parts[2]

	expectedSignature := GenerateHMACSHA256(header + "." + payload)
	if signature != expectedSignature {
		log.Println("Invalid token signature")
		log.Println("Expected signature:", expectedSignature)
		log.Println("Provided signature:", signature)
		return nil, fmt.Errorf("invalid token signature")
	}

	payloadData, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		log.Println("Invalid token payload:", err)
		return nil, fmt.Errorf("invalid token payload")
	}

	var claims Claims
	err = json.Unmarshal(payloadData, &claims)
	if err != nil {
		log.Println("Invalid token claims:", err)
		return nil, fmt.Errorf("invalid token claims")
	}

	if time.Now().Unix() > claims.Exp {
		log.Println("Token has expired")
		return nil, fmt.Errorf("token has expired")
	}

	log.Println("Token is valid, UserID:", claims.UserID)
	return &claims, nil
}

func (s *MyServer) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// ignore la vérification pour les requêtes OPTIONS (CORS) et la route de déconnexion
		if r.Method == http.MethodOptions || r.URL.Path == "/logout" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("No Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Println("Token reçu :", authHeader)

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Println("Invalid Authorization format")
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]
		claims, err := VerifyJWT(token)
		if err != nil {
			log.Println("Token verification failed:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// injecte l'ID utilisateur dans le contexte
		log.Println("User ID from token:", claims.UserID)
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (s *MyServer) VerifyTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// extraire le token de l'en-tête Authorization
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]
		claims, err := VerifyJWT(token)
		if err != nil {
			log.Println("Token verification failed:", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Si le token est valide, répondre avec un statut 200 ok
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Token is valid. User ID: %s", claims.UserID)))
	}
}
*/
