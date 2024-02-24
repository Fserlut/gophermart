package models

type User struct {
	UUID     string
	Login    string
	Password string
}

type UserRegisterOrLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
