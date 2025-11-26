package domain

type User struct {
	ID       string
	Username string
	IsActive bool
	TeamName string
}

type ActivateUserRequest struct {
	UserID   string
	IsActive bool
}
