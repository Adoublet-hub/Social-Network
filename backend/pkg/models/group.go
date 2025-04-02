package models

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
)

// structure de base d'un groupe
type Group struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name" validate:"required"`
	Description string        `json:"description" validate:"required"`
	CreatorID   uuid.UUID     `json:"creator_id"`
	Members     []GroupMember `json:"members,omitempty"`
	Events      []GroupEvent  `json:"events,omitempty"` // Liste des événements
	CreatedAt   string        `json:"created_at"`
}

// structure pour les membres du groupe
type GroupMember struct {
	UserID   string         `json:"user_id"`  // ID utilisateur
	Username string         `json:"username"` // Nom de l'utilisateur
	Avatar   sql.NullString `json:"image_profil,omitempty"`
	Role     string         `json:"role"`   // Rôle dans le groupe (creator/member)
	Status   string         `json:"status"` // Statut (pending/accepted)
}

type Message struct {
	ID             uuid.UUID `json:"id" validate:"required"`
	SenderUsername string    `json:"sender_username" validate:"required"`
	TargetUsername string    `json:"target_username" validate:"required"`
	Content        string    `json:"content"`
	Timestamp      time.Time `json:"timestamp"`
	Type           string    `json:"type"`
	Emoji          string    `json:"emoji,omitempty"`
}
