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
