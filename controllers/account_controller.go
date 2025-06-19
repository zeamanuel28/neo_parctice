package controllers

import (
	"encoding/json"
	"neobank-lite/database"
	"neobank-lite/middleware"
	"neobank-lite/models"
	"net/http"
	"strconv" // Added for strconv.Atoi

	"github.com/google/uuid"
	"gorm.io/gorm" // Added for gorm.ErrRecordNotFound
)

type CreateAccountRequest struct {
	Balance      float64 `json:"balance"`
	account_type string  // This field IS included, determined by user input
}

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr) // POTENTIAL ISSUE: Error from Atoi is ignored

	// Parse the request body for the initial balance
	var req CreateAccountRequest // Declare req here
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// VALIDATION: Check for negative balance
	if req.Balance < 0 {
		http.Error(w, "Initial balance cannot be negative", http.StatusBadRequest)
		return
	}

	// Check if user already has an account
	var existing models.Account
	result := database.DB.Where("user_id = ?", userID).First(&existing)
	if result.Error == nil { // If result.Error is nil, a record WAS found
		http.Error(w, "User already has an account", http.StatusBadRequest)
		return
	}
	// POTENTIAL ISSUE: No explicit check for other database errors here.
	// If result.Error is NOT gorm.ErrRecordNotFound, but some other DB error,
	// the code will proceed and likely fail on the Create operation.
	// It's better to explicitly check for `gorm.ErrRecordNotFound` and
	// handle other errors as internal server errors.

	account := models.Account{
		AccountNumber: uuid.New().String(),
		UserID:        userID,
		Balance:       req.Balance,      // Balance now comes from req.Balance
		AccountType:   req.account_type, // AccountType is hardcoded to "virtual"
	}

	if err := database.DB.Create(&account).Error; err != nil {
		// Log the error for debugging purposes
		// log.Printf("Error creating account for user %d: %v", userID, err)
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json") // Good practice to set content type
	w.WriteHeader(http.StatusCreated)                  // Set 201 Created status for successful creation
	json.NewEncoder(w).Encode(account)
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr) // POTENTIAL ISSUE: Error from Atoi is ignored

	var account models.Account
	result := database.DB.Where("user_id = ?", userID).First(&account)
	if result.Error == gorm.ErrRecordNotFound {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}
	// POTENTIAL ISSUE: No explicit check for other database errors.
	// If result.Error is not ErrRecordNotFound but some other DB error,
	// the code will proceed with an empty 'account' and likely return
	// default values (0 for balance, empty string for account_number),
	// or panic if you try to access other fields assuming it's valid.
	if result.Error != nil { // Handle other DB errors for GetBalance
		// log.Printf("Database error getting balance for user %d: %v", userID, result.Error)
		http.Error(w, "Failed to retrieve account balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json") // Good practice to set content type
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_number": account.AccountNumber,
		"balance":        account.Balance,
	})
}
