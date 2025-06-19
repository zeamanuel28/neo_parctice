package models

type Account struct {
	AccountNumber string  `json:"account_number" gorm:"primaryKey"` // Separate tags with a space
	UserID        int     `json:"user_id"`
	Balance       float64 `json:"balance"`
	AccountType   string  `json:"account_type"` // savings, virtual
}
