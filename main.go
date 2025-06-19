package main

import (
	"fmt"
	"log"
	"neobank-lite/config"
	"neobank-lite/database"
	"neobank-lite/routes"
	"net/http"
)

func main() {
	config.LoadEnvVariables()
	database.Connect()

	router := routes.SetupRouter()

	fmt.Println("🚀 Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
