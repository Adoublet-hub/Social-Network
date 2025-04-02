package controllers

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
)

func GetUserIDbyEmail(db *sql.DB, email string) (uuid.UUID, error) {
	userID := uuid.Must(uuid.NewV4())
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("email incorrect")
	}
	return userID, nil
}

func GetUsernameByEmail(db *sql.DB, email string) (string, error) {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE email = ?", email).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func GetPasswordByEmail(db *sql.DB, email string) (string, error) {
	var passWordId string
	err := db.QueryRow("SELECT password_hash FROM users WHERE email = ?", email).Scan(&passWordId)
	if err != nil {
		return "", err
	}
	return passWordId, nil
}

func GetUsernameByID(db *sql.DB, userID uuid.UUID) (string, error) {
	var username string
	err := db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func GetUserIDbyUsername(db *sql.DB, username string) (uuid.UUID, error) {
	userID := uuid.Must(uuid.NewV4())
	query := "SELECT id FROM users WHERE username = ?"
	err := db.QueryRow(query, username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No user found with username:", username)
		}
		return uuid.Nil, fmt.Errorf("failed to get user ID by username: %w", err)
	}
	return userID, nil
}

func GetPasswordByUsername(db *sql.DB, username string) (string, error) {
	var password string
	query := "SELECT password_hash FROM users WHERE username = ?"
	err := db.QueryRow(query, username).Scan(&password)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No password found for username:", username)
		}
		return "", fmt.Errorf("failed to get password by username: %w", err)
	}
	return password, nil
}
