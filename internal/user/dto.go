package user

import "time"

type Friend struct {
	ID          int
	UserID      int
	FriendID    int
	Status      string
	FriendSince time.Time
}
