package dto

type CreateAccountRequest struct {
	Balance     float64 `json:"balance"`
	AccountType string  `json:"account_type"`
	PhoneNumber int     `json:"phone_number"`
}
