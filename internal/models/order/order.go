package order

import "time"

type Order struct {
	Number      string    `json:"number"`
	UserUUID    string    `json:"-"`
	Status      string    `json:"status"`
	Accrual     *float64  `json:"accrual,omitempty"`
	Withdraw    *float64  `json:"withdraw,omitempty"`
	ProcessedAt time.Time `json:"processed_at,omitempty"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type OrderInfo struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
