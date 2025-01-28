package user

type FriendStore interface {
	CreateFriendship(userID, recipientID int) (FriendshipResponse, error)
	UpdateFriendship(friendshipID int, status string) (FriendshipResponse, error)
	GetAllFriendships(userID int, status string) ([]FriendshipResponse, error)
}
