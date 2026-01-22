package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	OpenAIAPIKey string
	DBPath       string
	ChunkSize    int
	Overlap      int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		DBPath:       getEnvOrDefault("DB_PATH", "db_data/doc_search.db"),
		ChunkSize:    getEnvAsIntOrDefault("CHUNK_SIZE", 1000),
		Overlap:      getEnvAsIntOrDefault("OVERLAP", 100),
	}

	// Validate required fields
	if cfg.OpenAIAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Validate chunk size and overlap
	if cfg.ChunkSize <= 0 {
		return nil, fmt.Errorf("CHUNK_SIZE must be positive, got %d", cfg.ChunkSize)
	}
	if cfg.Overlap < 0 {
		return nil, fmt.Errorf("OVERLAP must be non-negative, got %d", cfg.Overlap)
	}
	if cfg.Overlap >= cfg.ChunkSize {
		return nil, fmt.Errorf("OVERLAP must be less than CHUNK_SIZE (overlap: %d, chunk_size: %d)", cfg.Overlap, cfg.ChunkSize)
	}

	return cfg, nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsIntOrDefault returns environment variable as int or default
func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
