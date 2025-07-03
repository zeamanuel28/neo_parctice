package controllers

import (
	"encoding/json"
	"neobank-lite/database"
	"neobank-lite/dto"
	"neobank-lite/models"
	"neobank-lite/utils"
	"net/http"

	"gorm.io/gorm"
)

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email, password, and national ID image
// @Tags Auth
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Full Name"
// @Param email formData string true "Email Address"
// @Param password formData string true "Password"
// @Param national_id formData file true "National ID Image"
// @Success 201 {object} map[string]string
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	// Limit the size to 10MB
	r.ParseMultipartForm(10 << 20)

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Get uploaded file
	file, handler, err := r.FormFile("national_id")
	if err != nil {
		http.Error(w, "Failed to read national ID image", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate image path (store inside uploads/national_ids/)
	imagePath := "./uploads/national_ids/" + handler.Filename

	// Save the file to disk
	err = utils.CreateImageFile(imagePath, file)
	if err != nil {
		http.Error(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user
	user := models.User{
		Name:       name,
		Email:      email,
		Password:   hashedPassword,
		NationalID: imagePath,
		KYCStatus:  "pending",
		Role:       "user",
	}

	if result := database.DB.Create(&user); result.Error != nil {
		http.Error(w, "User already exists or invalid data", http.StatusBadRequest)
		return
	}

	// Success
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully!"})
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Router /login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	//var input models.User
	var req dto.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid login data", http.StatusBadRequest)
		return
	}

	var user models.User
	result := database.DB.Where("email = ?", req.Email).First(&user)
	if result.Error == gorm.ErrRecordNotFound {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		http.Error(w, "Incorrect password", http.StatusUnauthorized)
		return
	}

	token, _ := utils.GenerateJWT(user.ID, user.Role)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
