package models

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	AccountNumber string  `json:"account_number" gorm:"primaryKey"` // Separate tags with a space
	UserID        int     `json:"user_id"`
	Balance       float64 `json:"balance"`
	AccountType   string  `json:"account_type"`
	PhoneNumber   int     `json:"phone_number"`

	// savings, virtual
}
