package model

type User struct {
	ID           uint64
	Login        string
	PasswordHash string
}

type RegisterUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
