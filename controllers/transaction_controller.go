package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"neobank-lite/database"
	"neobank-lite/dto"
	"neobank-lite/middleware"
	"neobank-lite/models"
)

var (
	transactionChan = make(chan TransactionJob, 100) // buffered channel
	mu              sync.Mutex                       // mutex for safe access
)

type TransactionJob struct {
	Type      string
	UserID    int
	Amount    float64
	ToAccount string
	Response  chan error
}

func init() {
	go processTransactions()
}

func processTransactions() {
	for job := range transactionChan {
		switch job.Type {
		case "deposit":
			err := handleDeposit(job.UserID, job.Amount)
			job.Response <- err
		case "transfer":
			err := handleTransfer(job.UserID, job.ToAccount, job.Amount)
			job.Response <- err
		}
	}
}

func handleDeposit(userID int, amount float64) error {
	mu.Lock()
	defer mu.Unlock()

	var account models.Account
	if err := database.DB.First(&account, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("account not found")
	}
	tx := database.DB.Begin()
	account.Balance += amount
	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		return err
	}
	transaction := models.Transaction{
		FromAccount: account.AccountNumber,
		ToAccount:   account.AccountNumber,
		Amount:      amount,
		Type:        "deposit",
		Timestamp:   time.Now(),
		Status:      "success",
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func handleTransfer(userID int, toAccount string, amount float64) error {
	mu.Lock()
	defer mu.Unlock()

	var sender models.Account
	if err := database.DB.First(&sender, "user_id = ?", userID).Error; err != nil {
		return fmt.Errorf("sender account not found")
	}

	var receiver models.Account
	if err := database.DB.First(&receiver, "account_number = ?", toAccount).Error; err != nil {
		return fmt.Errorf("receiver account not found")
	}

	if sender.Balance < amount {
		return fmt.Errorf("insufficient funds")
	}
	tx := database.DB.Begin()
	sender.Balance -= amount
	receiver.Balance += amount
	if err := tx.Save(&sender).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Save(&receiver).Error; err != nil {
		tx.Rollback()
		return err
	}
	transaction := models.Transaction{
		FromAccount: sender.AccountNumber,
		ToAccount:   receiver.AccountNumber,
		Amount:      amount,
		Type:        "transfer",
		Timestamp:   time.Now(),
		Status:      "success",
	}
	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func Deposit(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found.", http.StatusInternalServerError)
		return
	}
	if user.KYCStatus != "verified" {
		http.Error(w, "KYC not verified", http.StatusForbidden)
		return
	}

	var req dto.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	respChan := make(chan error)
	transactionChan <- TransactionJob{
		Type:     "deposit",
		UserID:   userID,
		Amount:   req.Amount,
		Response: respChan,
	}
	err := <-respChan
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Deposit successful",
	})
}

type TransferRequest struct {
	ToAccount string  `json:"to_account"`
	Amount    float64 `json:"amount"`
}

func Transfer(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}
	if user.KYCStatus != "verified" {
		http.Error(w, "KYC not verified", http.StatusForbidden)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		http.Error(w, "Invalid transfer data", http.StatusBadRequest)
		return
	}

	respChan := make(chan error)
	transactionChan <- TransactionJob{
		Type:      "transfer",
		UserID:    userID,
		ToAccount: req.ToAccount,
		Amount:    req.Amount,
		Response:  respChan,
	}
	err := <-respChan
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transfer successful",
	})
}

func TransactionHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var account models.Account
	if err := database.DB.First(&account, "user_id = ?", userID).Error; err != nil {
		http.Error(w, "Account not found for user.", http.StatusInternalServerError)
		return
	}

	var transactions []models.Transaction
	if err := database.DB.Where("from_account = ? OR to_account = ?", account.AccountNumber, account.AccountNumber).Order("timestamp desc").Find(&transactions).Error; err != nil {
		http.Error(w, "Failed to retrieve transaction history.", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transactions)
}
