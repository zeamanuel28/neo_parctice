package controllers

import (
	"encoding/json"
	"neobank-lite/database"
	"neobank-lite/middleware"
	"neobank-lite/models"
	"net/http"
	"strconv"
)

type KYCRequest struct {
	NationalID string `json:"national_id"`
}

// SubmitKYC godoc
// @Summary Submit KYC verification
// @Description Verifies the user KYC status using their registered national ID
// @Tags KYC
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /kyc/verify [post]
func SubmitKYC(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var user models.User
	database.DB.First(&user, userID)

	if user.NationalID == "" {
		http.Error(w, "National ID not found in profile. Please update your profile.", http.StatusBadRequest)
		return
	}

	// In real banking, we might verify this national ID via 3rd-party API here

	user.KYCStatus = "verified"
	database.DB.Save(&user)

	json.NewEncoder(w).Encode(map[string]string{
		"message":    "KYC verified using registered national ID",
		"kyc_status": user.KYCStatus,
	})
}

// GetKYCStatus godoc
// @Summary Get current KYC status
// @Description Returns the current KYC verification status of the user
// @Tags KYC
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {string} string "Internal Server Error"
// @Router /kyc/status [get]
func GetKYCStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := middleware.GetUserIDFromContext(r)
	userID, _ := strconv.Atoi(userIDStr)

	var user models.User
	database.DB.First(&user, userID)

	json.NewEncoder(w).Encode(map[string]string{
		"kyc_status":  user.KYCStatus,
		"national_id": user.NationalID,
	})
}
