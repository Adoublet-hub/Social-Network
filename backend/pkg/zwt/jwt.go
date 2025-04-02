package zwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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
