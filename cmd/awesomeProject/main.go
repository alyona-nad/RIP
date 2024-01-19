package main

import (
	"log"

	app "awesomeProject/internal/api"
)

// @title Производство красок
// @version 1.0
// @description Colorants app

// @host localhost:8080
// @BasePath /

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

