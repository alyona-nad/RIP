package main

import (
	"log"

	app "awesomeProject/internal/api"
)

// @title Dyes from Colorants
// @version 1.0
// @description colorants app

// @host localhost:8080
// @BasePath /RIP_frontend/

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	log.Println("Application start!")
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	application.StartServer()
	log.Println("Application terminated!")

}
