package controllers

import (
	"encoding/json"
	"neobank-lite/database"
	"neobank-lite/dto"
	"neobank-lite/middleware"
	"neobank-lite/models" // Assuming your User model is here
	"net/http"
	"strconv"
	"time"
)

// Deposit request structure

var req dto.DepositRequest

func Deposit(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	// --- KYC Status Check Start ---
	var user models.User // Assuming your User model is in models package
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found.", http.StatusInternalServerError)
		return
	}

	if user.KYCStatus != "verified" { // Check if KYC status is "verified"
		http.Error(w, "KYC status not verified. Deposit not allowed.", http.StatusForbidden) // Use StatusForbidden for permission issues
		return
	}
	// --- KYC Status Check End ---

	//var req DepositRequest.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid deposit amount", http.StatusBadRequest)
		return
	}

	var account models.Account
	// It's good practice to check if the account exists
	if err := database.DB.First(&account, "user_id = ?", userID).Error; err != nil {
		http.Error(w, "Account not found for user.", http.StatusInternalServerError)
		return
	}

	// Start a database transaction for atomicity
	txDB := database.DB.Begin()
	if txDB.Error != nil {
		http.Error(w, "Failed to start database transaction", http.StatusInternalServerError)
		return
	}

	// Update account balance
	account.Balance += req.Amount
	if err := txDB.Save(&account).Error; err != nil {
		txDB.Rollback() // Rollback on error
		http.Error(w, "Failed to update account balance.", http.StatusInternalServerError)
		return
	}

	// Log transaction
	transactionRecord := models.Transaction{
		FromAccount: account.AccountNumber, // For a deposit, FromAccount is usually the same as ToAccount or empty
		ToAccount:   account.AccountNumber,
		Amount:      req.Amount,
		Timestamp:   time.Now(),
		Type:        "deposit",
		Status:      "success", // Assume success unless an error occurs later
	}
	if err := txDB.Create(&transactionRecord).Error; err != nil {
		txDB.Rollback() // Rollback on error
		http.Error(w, "Failed to log transaction.", http.StatusInternalServerError)
		return
	}

	// Commit the transaction if all operations were successful
	txDB.Commit()
	if txDB.Error != nil {
		http.Error(w, "Failed to commit database transaction.", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Deposit successful", "new_balance": strconv.FormatFloat(account.Balance, 'f', 2, 64)})
}

// Transfer request structure (No change needed for KYC here, as sender's KYC would be checked before any operation)
type TransferRequest struct {
	ToAccount string  `json:"to_account"`
	Amount    float64 `json:"amount"`
}

func Transfer(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	// --- KYC Status Check for Sender in Transfer (Recommended) ---
	var senderUser models.User
	if err := database.DB.First(&senderUser, userID).Error; err != nil {
		http.Error(w, "User not found.", http.StatusInternalServerError)
		return
	}

	if senderUser.KYCStatus != "verified" {
		http.Error(w, "Your KYC status is not verified. Transfer not allowed.", http.StatusForbidden)
		return
	}
	// --- End KYC Status Check for Sender ---

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid transfer data", http.StatusBadRequest)
		return
	}

	var sender models.Account
	if err := database.DB.First(&sender, "user_id = ?", userID).Error; err != nil {
		http.Error(w, "Sender account not found.", http.StatusInternalServerError)
		return
	}

	if sender.Balance < req.Amount {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}

	var receiver models.Account
	if err := database.DB.First(&receiver, "account_number = ?", req.ToAccount).Error; err != nil {
		http.Error(w, "Recipient account not found", http.StatusBadRequest)
		return
	}

	// Start a database transaction for atomicity in transfers
	txDB := database.DB.Begin()
	if txDB.Error != nil {
		http.Error(w, "Failed to start database transaction", http.StatusInternalServerError)
		return
	}

	// Perform transfer
	sender.Balance -= req.Amount
	receiver.Balance += req.Amount

	if err := txDB.Save(&sender).Error; err != nil {
		txDB.Rollback()
		http.Error(w, "Failed to update sender account.", http.StatusInternalServerError)
		return
	}
	if err := txDB.Save(&receiver).Error; err != nil {
		txDB.Rollback()
		http.Error(w, "Failed to update receiver account.", http.StatusInternalServerError)
		return
	}

	transactionRecord := models.Transaction{
		FromAccount: sender.AccountNumber,
		ToAccount:   receiver.AccountNumber,
		Amount:      req.Amount,
		Timestamp:   time.Now(),
		Type:        "transfer",
		Status:      "success",
	}
	if err := txDB.Create(&transactionRecord).Error; err != nil {
		txDB.Rollback()
		http.Error(w, "Failed to log transaction.", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	txDB.Commit()
	if txDB.Error != nil {
		http.Error(w, "Failed to commit database transaction.", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Transfer successful"})
}

func TransactionHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var account models.Account
	// Check if the account exists
	if err := database.DB.First(&account, "user_id = ?", userID).Error; err != nil {
		http.Error(w, "Account not found for user.", http.StatusInternalServerError)
		return
	}

	var transactions []models.Transaction
	// Find transactions related to this account (either as sender or receiver)
	if err := database.DB.Where("from_account = ? OR to_account = ?", account.AccountNumber, account.AccountNumber).Order("timestamp desc").Find(&transactions).Error; err != nil {
		http.Error(w, "Failed to retrieve transaction history.", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transactions)
}
