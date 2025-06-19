package controllers

import (
	"encoding/json"
	"neobank-lite/database"
	"neobank-lite/models"
	"neobank-lite/utils"
	"net/http"

	"gorm.io/gorm"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var input models.User
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	hashed, _ := utils.HashPassword(input.Password)
	input.Password = hashed

	if result := database.DB.Create(&input); result.Error != nil {
		http.Error(w, "User already exists or invalid data", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully!"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input models.User
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid login data", http.StatusBadRequest)
		return
	}

	var user models.User
	result := database.DB.Where("email = ?", input.Email).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	token, _ := utils.GenerateJWT(user.ID)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
