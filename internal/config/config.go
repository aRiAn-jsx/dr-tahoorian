package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	Env          string
	DBType       string
	DBDSN        string
	SessionSecret string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		Port:         port,
		Env:          os.Getenv("ENV"),
		DBType:       os.Getenv("DB_TYPE"),
		DBDSN:        os.Getenv("DB_DSN"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
	}
}
