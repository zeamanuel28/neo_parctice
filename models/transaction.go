package models

import "time"

type Transaction struct {
	ID          int       `json:"id"`
	FromAccount string    `json:"from_account"`
	ToAccount   string    `json:"to_account"`
	Amount      float64   `json:"amount"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`   // deposit, withdraw, transfer
	Status      string    `json:"status"` // success, failed
}
