package main

import (
	"log"

	"github.com/JennerWork/chatbot/internal/app"
)

func main() {
	configPath := "../../config.yaml"
	if err := app.Run(configPath); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
