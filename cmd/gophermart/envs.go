package main

import (
	"github.com/ShukinDmitriy/gophermart/internal/config"
	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
	"os"
)

// LoadEnvFile is invoked before main()
func LoadEnvFile() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func parseEnvs(config *config.Config) {
	runAddress, exists := os.LookupEnv("RUN_ADDRESS")
	if exists {
		config.RunAddress = runAddress
	}

	databaseURI, exists := os.LookupEnv("DATABASE_URI")
	if exists {
		config.DatabaseURI = databaseURI
	}

	accrualSystemAddress, exists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if exists {
		config.AccrualSystemAddress = accrualSystemAddress
	}

	jwtSecretKey, exists := os.LookupEnv("JWT_SECRET_KEY")
	if exists {
		config.JwtSecretKey = jwtSecretKey
	}
}
