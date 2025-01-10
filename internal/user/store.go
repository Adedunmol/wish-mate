package user

type Store interface {
	CreateUser(body *CreateUserBody) (CreateUserResponse, error)
	FindUserByEmail(email string) (User, error)
	ComparePasswords(storedPassword, candidatePassword string) bool
}
