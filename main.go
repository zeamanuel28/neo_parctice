package main

import (
	"fmt"
	"log"
	"net/http"

	"neobank-lite/config"
	"neobank-lite/database"
	"neobank-lite/routes"

	_ "neobank-lite/docs" // 👈 Required for Swagger docs

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           NeoBank Lite API
// @version         1.0
// @description     REST API for NeoBank Lite Project
// @termsOfService  http://neobank-lite.com/terms

// @contact.name   NeoBank Support
// @contact.url    http://neobank-lite.com/support
// @contact.email  support@neobank-lite.com

// @host      localhost:8080
// @BasePath  /
// @schemes   http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	config.LoadEnvVariables()
	database.Connect()

	router := routes.SetupRouter()

	// Serve Swagger docs at /swagger/index.html
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
