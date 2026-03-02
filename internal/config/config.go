// Package config handles application configuration from environment variables.
package config

import "os"

// Config holds the application configuration values.
type Config struct {
	Port             string
	QueryDatabaseURL string
	ChatDatabaseURL  string
	AI               AIConfig
	PublicKey    string
}

// AIConfig holds AI/LLM-related configuration.
type AIConfig struct {
	// APIKey is the API key for the OpenAI-compatible service.
	APIKey string
	// BaseURL is the base URL of the OpenAI-compatible API endpoint
	// (e.g. "https://api.openai.com/v1").
	BaseURL string
	// Model is the model identifier in "provider/model" format
	// (e.g. "openai-compat/gpt-4o-mini").
	Model string

}

// Load reads configuration from environment variables and returns a Config.
// It falls back to sensible defaults when variables are not set.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := os.Getenv("AI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	publicKey := os.Getenv("PUBLIC_KEY")

	return &Config{
		Port:             port,
		QueryDatabaseURL: os.Getenv("QUERY_DATABASE_URL"),
		ChatDatabaseURL:  os.Getenv("CHAT_DATABASE_URL"),
		AI: AIConfig{
			APIKey:  os.Getenv("OPENAI_API_KEY"),
			BaseURL: baseURL,
			Model:   model,
		},
		PublicKey: publicKey,
	}
}
