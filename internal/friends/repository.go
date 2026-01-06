package friends

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SendFriendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID) error {
	query := `
		INSERT INTO friendships (requester_id, addressee_id, status)
		VALUES ($1, $2, 'pending')
		ON CONFLICT (requester_id, addressee_id) DO NOTHING`

	_, err := r.db.ExecContext(ctx, query, requesterID, addresseeID)
	return err
}

func (r *Repository) AcceptFriendRequest(ctx context.Context, friendshipID uuid.UUID) error {
	query := `UPDATE friendships SET status = 'accepted', updated_at = NOW() WHERE id = $1 AND status = 'pending'`
	_, err := r.db.ExecContext(ctx, query, friendshipID)
	return err
}

func (r *Repository) DeclineFriendRequest(ctx context.Context, friendshipID uuid.UUID) error {
	query := `DELETE FROM friendships WHERE id = $1 AND status = 'pending'`
	_, err := r.db.ExecContext(ctx, query, friendshipID)
	return err
}

func (r *Repository) RemoveFriend(ctx context.Context, characterID, friendID uuid.UUID) error {
	query := `DELETE FROM friendships WHERE ((requester_id = $1 AND addressee_id = $2) OR (requester_id = $2 AND addressee_id = $1))`
	_, err := r.db.ExecContext(ctx, query, characterID, friendID)
	return err
}

func (r *Repository) GetFriends(ctx context.Context, characterID uuid.UUID) ([]FriendWithDetails, error) {
	var friends []FriendWithDetails
	query := `
		SELECT
			f.id as friendship_id,
			CASE WHEN f.requester_id = $1 THEN f.addressee_id ELSE f.requester_id END as character_id,
			c.name as character_name,
			c.house,
			c.grade,
			c.level,
			f.status,
			(f.requester_id = $1) as is_requester,
			c.last_played_at,
			f.created_at
		FROM friendships f
		JOIN characters c ON c.id = CASE WHEN f.requester_id = $1 THEN f.addressee_id ELSE f.requester_id END
		WHERE (f.requester_id = $1 OR f.addressee_id = $1) AND f.status = 'accepted'
		ORDER BY c.last_played_at DESC NULLS LAST`

	if err := r.db.SelectContext(ctx, &friends, query, characterID); err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}

	return friends, nil
}

func (r *Repository) GetPendingRequests(ctx context.Context, characterID uuid.UUID) ([]FriendWithDetails, error) {
	var requests []FriendWithDetails
	query := `
		SELECT
			f.id as friendship_id,
			f.requester_id as character_id,
			c.name as character_name,
			c.house,
			c.grade,
			c.level,
			f.status,
			false as is_requester,
			c.last_played_at,
			f.created_at
		FROM friendships f
		JOIN characters c ON c.id = f.requester_id
		WHERE f.addressee_id = $1 AND f.status = 'pending'
		ORDER BY f.created_at DESC`

	if err := r.db.SelectContext(ctx, &requests, query, characterID); err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}

	return requests, nil
}

func (r *Repository) AreFriends(ctx context.Context, char1ID, char2ID uuid.UUID) (bool, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM friendships
		WHERE ((requester_id = $1 AND addressee_id = $2) OR (requester_id = $2 AND addressee_id = $1))
		AND status = 'accepted'`

	if err := r.db.GetContext(ctx, &count, query, char1ID, char2ID); err != nil {
		return false, err
	}

	return count > 0, nil
}
