package config

import (
	"os"
	"strings"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	APIPort     string
	DBPath      string
	Env         string
	CORSOrigins []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		APIPort:     getEnv("API_PORT", "8080"),
		DBPath:      getEnv("DB_PATH", "data.db"),
		Env:         getEnv("ENV", "development"),
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "*"), ","),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
