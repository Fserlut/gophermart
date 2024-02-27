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
	Number      string    `json:"number"`
	UserUuid    string    `json:"-"`
	Status      string    `json:"status"`
	Accrual     *float64  `json:"accrual,omitempty"`
	Withdraw    *float64  `json:"withdraw,omitempty"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type UserBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawRequest struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}

type WithdrawalsResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type OrderInfo struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
