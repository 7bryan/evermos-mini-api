package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const minJWTSecretLength = 32

type Config struct {
	App AppConfig
	DB  DatabaseConfig
	JWT JWTConfig
}

type AppConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	var problems []string

	cfg := &Config{
		App: AppConfig{Port: getEnv("APP_PORT", "8000")},
		DB: DatabaseConfig{
			Host:     getEnv("DB_HOST", "127.0.0.1"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name:     os.Getenv("DB_NAME"),
		},
	}

	if _, err := strconv.Atoi(cfg.App.Port); err != nil {
		problems = append(problems, "APP_PORT must be a number")
	}
	if cfg.DB.User == "" {
		problems = append(problems, "DB_USER is required")
	}
	if cfg.DB.Name == "" {
		problems = append(problems, "DB_NAME is required")
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < minJWTSecretLength {
		problems = append(problems, fmt.Sprintf("JWT_SECRET is required and must be at least %d characters", minJWTSecretLength))
	}

	hours, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil || hours <= 0 {
		problems = append(problems, "JWT_EXPIRY_HOURS must be a positive integer")
		hours = 24
	}
	cfg.JWT = JWTConfig{Secret: secret, Expiry: time.Duration(hours) * time.Hour}

	if len(problems) > 0 {
		return nil, errors.New("invalid configuration: " + strings.Join(problems, "; "))
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
