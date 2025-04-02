package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
)

// User model with helper types for JSON unmarshaling
type User struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Age         int        `json:"age"`
	Email       string     `json:"email"`
	Password    string     `json:"password"`
	FirstName   NullString `json:"firstName"`
	LastName    NullString `json:"lastName"`
	Role        string     `json:"role"`
	Gender      string     `json:"gender"`
	DateOfBirth NullTime   `json:"dateOfBirth"`
	Avatar      NullString `json:"avatar"`
	Bio         NullString `json:"bio"`
	PhoneNumber NullString `json:"phoneNumber"`
	Address     NullString `json:"address"`
	IsPrivate   bool       `json:"isPrivate"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type NullString struct {
	sql.NullString
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil { // Si la chaîne n'est pas nulle
		ns.Valid = true
		ns.String = *s
	} else { // Si la chaîne est nulle ou vide
		ns.Valid = false
		ns.String = ""
	}
	return nil
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s != nil && *s != "" { // Vérifie que la chaîne n'est pas vide
		parsedTime, err := time.Parse(time.RFC3339, *s)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
		nt.Valid = true
		nt.Time = parsedTime
	} else { // Si la chaîne est vide ou nulle
		nt.Valid = false
		nt.Time = time.Time{}
	}
	return nil
}

// NullTime is a helper type to handle nullable dates
type NullTime struct {
	sql.NullTime
}

type Response struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}
