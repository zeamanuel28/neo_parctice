package routes

import (
	"neobank-lite/controllers"
	"neobank-lite/middleware"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("NeoBank API is up and running! 🚀"))
	}).Methods("GET")
	router.HandleFunc("/register", controllers.Register).Methods("POST")
	router.HandleFunc("/login", controllers.Login).Methods("POST")

	// Protected routes group
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.JWTAuth)

	// Example protected endpoint
	protected.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r)
		w.Write([]byte("✅ Authenticated. Your user ID: " + userID))

	}).Methods("GET")
	protected.HandleFunc("/account/create", controllers.CreateAccount).Methods("POST")
	protected.HandleFunc("/account/balance", controllers.GetBalance).Methods("GET")
	protected.HandleFunc("/transaction/deposit", controllers.Deposit).Methods("POST")
	protected.HandleFunc("/transaction/transfer", controllers.Transfer).Methods("POST")
	protected.HandleFunc("/transaction/history", controllers.TransactionHistory).Methods("GET")

	return router
}
