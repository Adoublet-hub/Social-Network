package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type UserProfil struct {
	UserID      uuid.UUID    `json:"user_id"`
	Username    string       `json:"username"`
	FirstName   NullString   `json:"firstName"`
	LastName    NullString   `json:"lastName"`
	Email       string       `json:"email"`
	Gender      string       `json:"gender"`
	Bio         NullString   `json:"bio"`
	IsPrivate   bool         `json:"is_private"`
	Avatar      NullString   `json:"image_profil,omitempty"`
	PhoneNumber NullString   `json:"phoneNumber"`
	Followers   []SimpleUser `json:"followers,omitempty"`
	Following   []SimpleUser `json:"following,omitempty"`
	Posts       []Post       `json:"posts,omitempty"`
	Role        string       `json:"role"`
	Address     NullString   `json:"address"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type SimpleUser struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
}

// alias pour struct temporaire pour encodage JSON
type UserProfilJSON struct {
	UserID         uuid.UUID    `json:"user_id"`
	Username       string       `json:"username"`
	FirstName      NullString   `json:"first_name"`
	LastName       NullString   `json:"last_name"`
	Bio            string       `json:"bio"`
	IsPrivate      bool         `json:"is_private"`
	Avatar         NullString   `json:"image_profil,omitempty"`
	FollowersCount int          `json:"followers_count"`
	FollowingCount int          `json:"following_count"`
	Followers      []SimpleUser `json:"followers,omitempty"`
	Following      []SimpleUser `json:"following,omitempty"`
	Posts          []Post       `json:"posts,omitempty"`
}
type UserProfilResponse struct {
	UserID      uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Email       string    `json:"email"`
	Gender      string    `json:"gender"`
	Bio         string    `json:"bio"`
	IsPrivate   bool      `json:"is_private"`
	Avatar      string    `json:"image_profil,omitempty"`
	PhoneNumber string    `json:"phoneNumber"`
	Address     string    `json:"address"`
	UpdatedAt   time.Time `json:"updated_at"`
	Posts       []Post    `json:"posts,omitempty"`
}
