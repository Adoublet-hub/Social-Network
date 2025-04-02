package models

import (
	"time"

	"github.com/gofrs/uuid"
)

type GroupEvent struct {
	ID          uuid.UUID    `json:"id"`
	GroupID     uuid.UUID    `json:"group_id"`
	UserID      uuid.UUID    `json:"user_id"`
	Title       string       `json:"title" validate:"required"`
	Description string       `json:"description" validate:"required"`
	EventDate   time.Time    `json:"event_date" validate:"required"`
	Options     EventOptions `json:"options"`
	CreatedAt   time.Time    `json:"created_at"`
}

type EventOptions struct {
	Going    int `json:"going"`
	NotGoing int `json:"not_going"`
}

type EventResponse struct {
	EventID  uuid.UUID `json:"event_id"`
	UserID   uuid.UUID `json:"user_id"`
	Response string    `json:"response" validate:"oneof=Going 'Not going'"`
}
