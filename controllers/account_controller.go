package controllers

import (
	"encoding/json"
	"neobank-lite/database"
	"neobank-lite/dto"
	"neobank-lite/middleware"
	"neobank-lite/models"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateAccountRequest struct {
	Balance     float64 `json:"balance"`
	AccountType string  `json:"account_type"`
	PhoneNumber int     `json:"phone_number"`
}

// CreateAccount godoc
// @Summary Create a new account
// @Description Allows a user to create a new bank account
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param account body dto.CreateAccountRequest true "Account creation data"
// @Success 201 {object} models.Account
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /account/create [post]
func CreateAccount(w http.ResponseWriter, r *http.Request) {

	userIDStr := middleware.GetUserIDFromContext(r)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID in context", http.StatusInternalServerError)
		return
	}

	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Balance < 0 {
		http.Error(w, "Initial balance cannot be negative", http.StatusBadRequest)
		return
	}

	// Check if phone number already has an account
	var existing models.Account
	result := database.DB.Where("phone_number = ?", req.PhoneNumber).First(&existing)

	if result.Error == nil {
		http.Error(w, "Phone number already has an account", http.StatusBadRequest)
		return
	} else if result.Error != gorm.ErrRecordNotFound {
		http.Error(w, "Database error while checking existing account", http.StatusInternalServerError)
		return
	}

	account := models.Account{
		AccountNumber: uuid.New().String(),
		UserID:        userID,
		Balance:       req.Balance,
		AccountType:   req.AccountType,
		PhoneNumber:   req.PhoneNumber,
	}

	if err := database.DB.Create(&account).Error; err != nil {
		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(account)
}

// GetBalance godoc
// @Summary Get account balance
// @Description Returns the account balance for the authenticated user
// @Tags Account
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {string} string "Account not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /account/balance [get]
func GetBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID in context", http.StatusInternalServerError)
		return
	}

	var account models.Account
	result := database.DB.Where("user_id = ?", userID).First(&account)

	if result.Error == gorm.ErrRecordNotFound {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	} else if result.Error != nil {
		http.Error(w, "Failed to retrieve account balance", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_number": account.AccountNumber,
		"balance":        account.Balance,
	})
}
