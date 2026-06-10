package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                  string
	Port                 string
	LogLevel             string
	LogSpaced            bool
	DatabaseURL          string
	JWTSecret            string
	JWTTTLMinutes        int
	AdminEmail           string
	AdminPasswordHash    string
	AILogRaw             bool
	UserTestCronSeconds  int
	UserTestCronEnabled  bool
	GraderURL            string
	GraderToken          string
	GraderTimeoutSec     int
	AWSS3Region          string
	AWSS3Bucket          string
	AWSS3AccessKeyID     string
	AWSS3SecretAccessKey string
	AWSS3Prefix          string
}

func Load() Config {
	_ = godotenv.Load()

	env := getEnv("ENV", "development")
	return Config{
		Env:                  env,
		Port:                 getEnv("PORT", "8080"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		LogSpaced:            getEnvBool("LOG_SPACED", isDevEnv(env)),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTTTLMinutes:        getEnvInt("JWT_TTL_MINUTES", 1440),
		AILogRaw:             getEnvBool("AI_LOG_RAW", false),
		UserTestCronSeconds:  getEnvInt("USER_TEST_CRON_SECONDS", 60),
		UserTestCronEnabled:  getEnvBool("USER_TEST_CRON_ENABLED", true),
		GraderURL:            getEnv("GRADER_URL", "http://localhost:8000"),
		GraderToken:          getEnv("GRADER_TOKEN", ""),
		GraderTimeoutSec:     getEnvInt("GRADER_TIMEOUT_SECONDS", 180),
		AWSS3Region:          getEnv("AWS_S3_REGION", ""),
		AWSS3Bucket:          getEnv("AWS_S3_BUCKET", ""),
		AWSS3AccessKeyID:     getEnv("AWS_S3_ACCESS_KEY_ID", ""),
		AWSS3SecretAccessKey: getEnv("AWS_S3_SECRET_ACCESS_KEY", ""),
		AWSS3Prefix:          getEnv("AWS_S3_PREFIX", ""),
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
