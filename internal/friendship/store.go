package friendship

import "github.com/jackc/pgx/v5"

type FriendStore interface {
	CreateFriendship(userID, recipientID int) (FriendshipResponse, error)
	UpdateFriendship(friendshipID int, status string) (FriendshipResponse, error)
	GetAllFriendships(userID int, status string) ([]FriendshipResponse, error)
	GetFriendship(requestID int) (FriendshipResponse, error)
}

type FriendshipStore struct {
	db *pgx.Conn
}

func NewFriendshipStore(db *pgx.Conn) *FriendshipStore {

	return &FriendshipStore{db: db}
}

func (f *FriendshipStore) CreateFriendship(userID, recipientID int) (FriendshipResponse, error) {
	return FriendshipResponse{}, nil
}

func (f *FriendshipStore) UpdateFriendship(friendshipID int, status string) (FriendshipResponse, error) {
	return FriendshipResponse{}, nil
}

func (f *FriendshipStore) GetAllFriendships(userID int, status string) ([]FriendshipResponse, error) {
	return nil, nil
}

func (f *FriendshipStore) GetFriendship(requestID int) (FriendshipResponse, error) {
	return FriendshipResponse{}, nil
}
