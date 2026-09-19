package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Environment string
}

func LoadConfig() *Config {
	// Carrega arquivo .env se existir
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] Arquivo .env não encontrado ou usando variáveis do ambiente do sistema.")
	}

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "postgres://financas_user:financas_password@localhost:5432/financas_sah?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "super_secret_hello_kitty_jwt_key_sah_2024")
	env := getEnv("ENVIRONMENT", "development")

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		Environment: env,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
