package app

import (
	"github.com/joho/godotenv"
	"log"
)

func Run() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables from system")
	}

	

}
