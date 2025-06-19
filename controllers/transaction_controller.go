package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"neobank-lite/database"
	"neobank-lite/middleware"
	"neobank-lite/models"
)

// Deposit request structure
type DepositRequest struct {
	Amount float64 `json:"amount"`
}

func Deposit(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid deposit amount", http.StatusBadRequest)
		return
	}

	var account models.Account
	database.DB.First(&account, "user_id = ?", userID)
	account.Balance += req.Amount
	database.DB.Save(&account)

	// Log transaction
	tx := models.Transaction{
		FromAccount: account.AccountNumber,
		ToAccount:   account.AccountNumber,
		Amount:      req.Amount,
		Timestamp:   time.Now(),
		Type:        "deposit",
		Status:      "success",
	}
	database.DB.Create(&tx)

	json.NewEncoder(w).Encode(map[string]string{"message": "Deposit successful"})
}

// Transfer request structure
type TransferRequest struct {
	ToAccount string  `json:"to_account"`
	Amount    float64 `json:"amount"`
}

func Transfer(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid transfer data", http.StatusBadRequest)
		return
	}

	var sender models.Account
	database.DB.First(&sender, "user_id = ?", userID)

	if sender.Balance < req.Amount {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	var receiver models.Account
	if err := database.DB.First(&receiver, "account_number = ?", req.ToAccount).Error; err != nil {
		http.Error(w, "Recipient account not found", http.StatusBadRequest)
		return
	}

	// Perform transfer
	sender.Balance -= req.Amount
	receiver.Balance += req.Amount
	database.DB.Save(&sender)
	database.DB.Save(&receiver)

	tx := models.Transaction{
		FromAccount: sender.AccountNumber,
		ToAccount:   receiver.AccountNumber,
		Amount:      req.Amount,
		Timestamp:   time.Now(),
		Type:        "transfer",
		Status:      "success",
	}
	database.DB.Create(&tx)

	json.NewEncoder(w).Encode(map[string]string{"message": "Transfer successful"})
}

func TransactionHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var account models.Account
	database.DB.First(&account, "user_id = ?", userID)

	var transactions []models.Transaction
	database.DB.Where("from_account = ? OR to_account = ?", account.AccountNumber, account.AccountNumber).Order("timestamp desc").Find(&transactions)

	json.NewEncoder(w).Encode(transactions)
}
