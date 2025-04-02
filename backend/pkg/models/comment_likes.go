package models

import "github.com/gofrs/uuid"

type CommentLike struct {
	CommentID uuid.UUID `json:"comment_id"`
	UserID    uuid.UUID `json:"user_id"`
}
