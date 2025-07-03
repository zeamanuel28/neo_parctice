package models

type Account struct {
	ID        uint    `json:"id" example:"1"`
	CreatedAt string  `json:"created_at" example:"2025-07-03T10:30:00Z"`
	UpdatedAt string  `json:"updated_at" example:"2025-07-03T10:30:00Z"`
	DeletedAt *string `json:"deleted_at,omitempty" example:"2025-07-03T10:30:00Z"`

	AccountNumber string  `json:"account_number" gorm:"primaryKey" example:"acc-123456"`
	UserID        int     `json:"user_id" example:"10"`
	Balance       float64 `json:"balance" example:"1500.50"`
	AccountType   string  `json:"account_type" example:"savings"`
	PhoneNumber   int     `json:"phone_number" example:"911234567"`
}
