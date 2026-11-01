package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	config := Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}

	if config.DBHost == "" ||
		config.DBPort == "" ||
		config.DBUser == "" ||
		config.DBPassword == "" ||
		config.DBName == "" {
		return nil, errors.New("missing required database configuration")
	}

	return &config, nil
}
