package friends

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SendFriendRequest(ctx context.Context, requesterID, addresseeID uuid.UUID) error {
	if requesterID == addresseeID {
		return fmt.Errorf("cannot send friend request to yourself")
	}

	return s.repo.SendFriendRequest(ctx, requesterID, addresseeID)
}

func (s *Service) AcceptFriendRequest(ctx context.Context, friendshipID uuid.UUID) error {
	return s.repo.AcceptFriendRequest(ctx, friendshipID)
}

func (s *Service) DeclineFriendRequest(ctx context.Context, friendshipID uuid.UUID) error {
	return s.repo.DeclineFriendRequest(ctx, friendshipID)
}

func (s *Service) RemoveFriend(ctx context.Context, characterID, friendID uuid.UUID) error {
	return s.repo.RemoveFriend(ctx, characterID, friendID)
}

func (s *Service) GetFriends(ctx context.Context, characterID uuid.UUID) ([]FriendWithDetails, error) {
	return s.repo.GetFriends(ctx, characterID)
}

func (s *Service) GetPendingRequests(ctx context.Context, characterID uuid.UUID) ([]FriendWithDetails, error) {
	return s.repo.GetPendingRequests(ctx, characterID)
}

func (s *Service) AreFriends(ctx context.Context, char1ID, char2ID uuid.UUID) (bool, error) {
	return s.repo.AreFriends(ctx, char1ID, char2ID)
}
