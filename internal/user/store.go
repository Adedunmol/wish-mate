package user

type FriendStore interface {
	CreateFriendship(userID, recipientID int) (interface{}, error)
}
