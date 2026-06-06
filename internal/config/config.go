package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                 string
	Port                string
	LogLevel            string
	LogSpaced           bool
	DatabaseURL         string
	JWTSecret           string
	JWTTTLMinutes       int
	AdminEmail          string
	AdminPasswordHash   string
	ChutesAIURL         string
	ChutesAIAPIKey      string
	DeepseekModel       string
	ModelRouting        string
	AILogRaw            bool
	WhisperBin          string
	WhisperModelPath    string
	WhisperLanguage     string
	UserTestCronSeconds int
	UserTestCronEnabled bool
}

func Load() Config {
	_ = godotenv.Load()

	env := getEnv("ENV", "development")
	return Config{
		Env:                 env,
		Port:                getEnv("PORT", "8080"),
		LogLevel:            getEnv("LOG_LEVEL", "info"),
		LogSpaced:           getEnvBool("LOG_SPACED", isDevEnv(env)),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		JWTSecret:           getEnv("JWT_SECRET", ""),
		JWTTTLMinutes:       getEnvInt("JWT_TTL_MINUTES", 1440),
		ChutesAIURL:         getEnv("CHUTES_AI_URL", ""),
		ChutesAIAPIKey:      getEnv("CHUTES_AI_API_KEY", ""),
		DeepseekModel:       getEnv("DEEPSEEK_MODEL", ""),
		ModelRouting:        getEnv("MODEL_ROUTING", ""),
		AILogRaw:            getEnvBool("AI_LOG_RAW", false),
		WhisperBin:          getEnv("WHISPER_BIN", "whisper-cli"),
		WhisperModelPath:    getEnv("WHISPER_MODEL_PATH", "models/ggml-small.bin"),
		WhisperLanguage:     getEnv("WHISPER_LANGUAGE", "en"),
		UserTestCronSeconds: getEnvInt("USER_TEST_CRON_SECONDS", 60),
		UserTestCronEnabled: getEnvBool("USER_TEST_CRON_ENABLED", true),
	}
}

func (c Config) IsDev() bool {
	return c.Env == "development" || c.Env == "dev" || c.Env == "local"
}

func isDevEnv(env string) bool {
	return env == "development" || env == "dev" || env == "local"
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
