package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env               string
	Port              string
	LogLevel          string
	DatabaseURL       string
	JWTSecret         string
	JWTTTLMinutes     int
	RunSeeder         bool
	AdminEmail        string
	AdminPasswordHash string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		Env:           getEnv("ENV", "development"),
		Port:          getEnv("PORT", "8080"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTTTLMinutes: getEnvInt("JWT_TTL_MINUTES", 1440),
		RunSeeder:     getEnvBool("RUN_SEEDER", false),
	}
}

func (c Config) IsDev() bool {
	return c.Env == "development" || c.Env == "dev" || c.Env == "local"
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}

	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
