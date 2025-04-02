package models

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

type Post struct {
	ID           uuid.UUID      `json:"id" validate:"required"`
	Title        string         `json:"title" validate:"required"`
	Category     string         `json:"category" validate:"required"`
	Content      string         `json:"content" validate:"required"`
	UserID       uuid.UUID      `json:"user_id" validate:"required"`
	Visibility   string         `json:"visibility" validate:"oneof=public private limited" default:"public"`
	CreatedAt    time.Time      `json:"created_at" default:"CURRENT_TIMESTAMP"`
	ImagePath    string         `json:"image_path,omitempty"`
	Username     string         `json:"username" validate:"required"`
	AllowedUsers []uuid.UUID    `json:"allowed_users,omitempty"`
	Avatar       sql.NullString `json:"image_profil,omitempty"`
	TotalLikes   int            `json:"total_likes"`
	LikedByUser  bool           `json:"liked_by_user"`
}

type PostGroup struct {
	ID        uuid.UUID `json:"id"`
	GroupID   uuid.UUID `json:"group_id"`
	UserID    uuid.UUID `json:"user_id"`
	Title     string    `json:"title" validate:"required"`
	Content   string    `json:"content" validate:"required"`
	Likes     int       `json:"likes"`    // Total des likes
	Username  string    `json:"username"` // Nom d'utilisateur de l'auteur
	Avatar    string    `json:"avatar"`   // Avatar de l'auteur
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
