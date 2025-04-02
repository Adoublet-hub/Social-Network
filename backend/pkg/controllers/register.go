package controllers

import (
	"backend/pkg/models"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (s MyServer) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}

		var user models.User
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Failed to read request body:", err)
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		log.Println("Request body received:", string(body))
		if err := json.Unmarshal(body, &user); err != nil {
			log.Println("Failed to decode request payload:", err)
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		DB, err := s.Store.OpenDatabase()
		if err != nil {
			log.Println("Failed to open database:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer DB.Close()

		tx, err := DB.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		defer func() {
			if err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		}()

		log.Println("Database and table ready for register")

		if err := RegisterUser(w, r, DB, user); err != nil {
			log.Println("Failed to create user:", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		var newUser models.User
		err = DB.QueryRow(`
			SELECT id, username, age, email, password_hash, first_name, last_name, role, gender, date_of_birth, avatar, bio, phone_number, address, is_private, created_at, updated_at 
			FROM users 
			WHERE email = ?`, user.Email).Scan(
			&newUser.ID,
			&newUser.Username,
			&newUser.Age,
			&newUser.Email,
			&newUser.Password,
			&newUser.FirstName,
			&newUser.LastName,
			&newUser.Role,
			&newUser.Gender,
			&newUser.DateOfBirth,
			&newUser.Avatar,
			&newUser.Bio,
			&newUser.PhoneNumber,
			&newUser.Address,
			&newUser.IsPrivate,
			&newUser.CreatedAt,
			&newUser.UpdatedAt,
		)

		if err != nil {
			log.Println("Failed to retrieve new user data:", err)
			http.Error(w, "Failed to retrieve user data", http.StatusInternalServerError)
			return
		}

		log.Println("Newly created user data:", newUser)

		response := models.Response{
			Message: "User registered successfully",
			User:    newUser,
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println("Failed to encode response:", err)
			http.Error(w, "Failed to send response", http.StatusInternalServerError)
			return
		}
	}
}

func RegisterUser(w http.ResponseWriter, r *http.Request, DB *sql.DB, user models.User) error {
	log.Println("Starting user registration process for email:", user.Email)

	if !IsValidEmail(user.Email) {
		log.Println("Invalid email format for:", user.Email)
		http.Error(w, "Invalid email format", http.StatusUnauthorized)
		return errors.New("invalid email format")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Failed to hash password:", err)
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	user.ID = uuid.Must(uuid.NewV4())

	if err = CreateUser(DB, user); err != nil {
		log.Println("Failed to insert user:", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return err
	}

	log.Println("User registration completed successfully for:", user.Email)
	return nil
}

func CreateUser(db *sql.DB, user models.User) error {
	var countEmail, countUsername int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", user.Email).Scan(&countEmail)
	if err != nil {
		log.Println("Failed to check email existence:", err)
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if countEmail > 0 {
		return fmt.Errorf("email already exists")
	}

	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", user.Username).Scan(&countUsername)
	if err != nil {
		log.Println("Failed to check username existence:", err)
		return fmt.Errorf("failed to check username existence: %w", err)
	}
	if countUsername > 0 {
		return fmt.Errorf("username already exists")
	}

	// execute insertion
	query := `INSERT INTO users 
        (id, username, age, email, password_hash, first_name, last_name, role, gender, date_of_birth, avatar, bio, phone_number, address, is_private, created_at, updated_at) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = db.Exec(query,
		user.ID,
		user.Username,
		user.Age,
		user.Email,
		user.Password,
		sql.NullString{String: user.FirstName.String, Valid: user.FirstName.Valid},
		sql.NullString{String: user.LastName.String, Valid: user.LastName.Valid},
		user.Role,
		user.Gender,
		sql.NullTime{Time: user.DateOfBirth.Time, Valid: user.DateOfBirth.Valid},
		sql.NullString{String: user.Avatar.String, Valid: user.Avatar.Valid},
		sql.NullString{String: user.Bio.String, Valid: user.Bio.Valid},
		sql.NullString{String: user.PhoneNumber.String, Valid: user.PhoneNumber.Valid},
		sql.NullString{String: user.Address.String, Valid: user.Address.Valid},
		user.IsPrivate,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		log.Println("Failed to execute insert query:", err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	log.Println("User successfully created with ID:", user.ID)
	return nil
}

func IsValidEmail(email string) bool {
	if email == "" {
		log.Println("Email validation failed: empty email")
		return false
	}
	at := strings.Index(email, "@")
	dot := strings.LastIndex(email, ".")
	isValid := at > 0 && dot > at+1 && dot < len(email)-1
	if !isValid {
		log.Println("Email validation failed for:", email)
	}
	return isValid
}
