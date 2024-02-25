package models

import (
	"time"
)

type User struct {
	UUID     string
	Login    string
	Password string
}

type UserRegisterOrLoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Order struct {
	Number     string    `json:"number"`
	UserUuid   string    `json:"-"`
	Status     string    `json:"status"`
	Accrual    *int64    `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}
