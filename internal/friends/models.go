package friends

import (
	"time"

	"github.com/google/uuid"
)

type FriendStatus string

const (
	StatusPending  FriendStatus = "pending"
	StatusAccepted FriendStatus = "accepted"
	StatusBlocked  FriendStatus = "blocked"
)

type Friendship struct {
	ID          uuid.UUID    `db:"id" json:"id"`
	RequesterID uuid.UUID    `db:"requester_id" json:"requester_id"`
	AddresseeID uuid.UUID    `db:"addressee_id" json:"addressee_id"`
	Status      FriendStatus `db:"status" json:"status"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
}

type FriendWithDetails struct {
	FriendshipID   uuid.UUID    `db:"friendship_id" json:"friendship_id"`
	CharacterID    uuid.UUID    `db:"character_id" json:"character_id"`
	CharacterName  string       `db:"character_name" json:"character_name"`
	House          string       `db:"house" json:"house"`
	Grade          int          `db:"grade" json:"grade"`
	Level          int          `db:"level" json:"level"`
	Status         FriendStatus `db:"status" json:"status"`
	IsRequester    bool         `db:"is_requester" json:"is_requester"`
	LastPlayedAt   *time.Time   `db:"last_played_at" json:"last_played_at,omitempty"`
	CreatedAt      time.Time    `db:"created_at" json:"created_at"`
}
