package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Address     string
	Port        string
	StoragePath string
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var cfg Config
	cfg.Address = os.Getenv("ADDRESS")
	cfg.Port = os.Getenv("PORT")
	cfg.StoragePath = os.Getenv("STORAGE_PATH")
	return &cfg
}
