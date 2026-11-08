package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	DatabaseUrl string
}

func MustLoad() Config {
	// Attempt to load .env if present (non-fatal if absent in Docker/prod)
	_ = godotenv.Load()

	port := strings.Trim(os.Getenv("PORT"), "\"'")
	if port == "" {
		port = "8080"
	}

	env := strings.Trim(os.Getenv("ENV"), "\"'")
	if env == "" {
		env = "development"
	}

	databaseUrl := strings.Trim(os.Getenv("DATABASE_URL"), "\"'")
	if databaseUrl == "" {
		panic("DATABASE_URL is required")
	}

	return Config{
		Port:        port,
		Env:         env,
		DatabaseUrl: databaseUrl,
	}
}
