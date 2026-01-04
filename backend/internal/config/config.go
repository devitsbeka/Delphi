package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	// Core
	Environment string
	LogLevel    string
	APIPort     int
	FrontendURL string

	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// Encryption
	EncryptionKey string

	// Supabase
	SupabaseURL            string
	SupabaseAnonKey        string
	SupabaseServiceRoleKey string

	// Fly.io
	FlyAPIToken string
	FlyOrg      string
	FlyRegion   string

	// GitHub
	GitHubAppID         string
	GitHubAppPrivateKey string
	GitHubClientID      string
	GitHubClientSecret  string
	GitHubWebhookSecret string

	// Stripe
	StripeSecretKey    string
	StripeWebhookSecret string
	StripePricePro      string
	StripePriceEnterprise string

	// AI Providers (default/fallback)
	OpenAIAPIKey    string
	AnthropicAPIKey string
	GoogleAIAPIKey  string
	OllamaBaseURL   string

	// Social Media
	TwitterAPIKey       string
	TwitterAPISecret    string
	LinkedInClientID    string
	LinkedInClientSecret string

	// Email
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	// Messaging
	SlackClientID     string
	SlackClientSecret string
	DiscordBotToken   string

	// Monitoring
	SentryDSN string
}

// Load reads configuration from environment variables and config files
func Load() (*Config, error) {
	v := viper.New()

	// Set config file options
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Read config file if exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Environment variables override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("API_PORT", 8080)
	v.SetDefault("FRONTEND_URL", "http://localhost:5173")
	v.SetDefault("REDIS_URL", "redis://localhost:6379")
	v.SetDefault("OLLAMA_BASE_URL", "http://localhost:11434")
	v.SetDefault("SMTP_PORT", 587)
	v.SetDefault("FLY_REGION", "iad")
	v.SetDefault("FLY_ORG", "personal")

	cfg := &Config{
		// Core
		Environment: v.GetString("ENVIRONMENT"),
		LogLevel:    v.GetString("LOG_LEVEL"),
		APIPort:     v.GetInt("API_PORT"),
		FrontendURL: v.GetString("FRONTEND_URL"),

		// Database
		DatabaseURL: v.GetString("DATABASE_URL"),

		// Redis
		RedisURL: v.GetString("REDIS_URL"),

		// Encryption
		EncryptionKey: v.GetString("ENCRYPTION_KEY"),

		// Supabase
		SupabaseURL:            v.GetString("SUPABASE_URL"),
		SupabaseAnonKey:        v.GetString("SUPABASE_ANON_KEY"),
		SupabaseServiceRoleKey: v.GetString("SUPABASE_SERVICE_ROLE_KEY"),

		// Fly.io
		FlyAPIToken: v.GetString("FLY_API_TOKEN"),
		FlyOrg:      v.GetString("FLY_ORG"),
		FlyRegion:   v.GetString("FLY_REGION"),

		// GitHub
		GitHubAppID:         v.GetString("GITHUB_APP_ID"),
		GitHubAppPrivateKey: v.GetString("GITHUB_APP_PRIVATE_KEY"),
		GitHubClientID:      v.GetString("GITHUB_CLIENT_ID"),
		GitHubClientSecret:  v.GetString("GITHUB_CLIENT_SECRET"),
		GitHubWebhookSecret: v.GetString("GITHUB_WEBHOOK_SECRET"),

		// Stripe
		StripeSecretKey:       v.GetString("STRIPE_SECRET_KEY"),
		StripeWebhookSecret:   v.GetString("STRIPE_WEBHOOK_SECRET"),
		StripePricePro:        v.GetString("STRIPE_PRICE_PRO"),
		StripePriceEnterprise: v.GetString("STRIPE_PRICE_ENTERPRISE"),

		// AI Providers
		OpenAIAPIKey:    v.GetString("OPENAI_API_KEY"),
		AnthropicAPIKey: v.GetString("ANTHROPIC_API_KEY"),
		GoogleAIAPIKey:  v.GetString("GOOGLE_AI_API_KEY"),
		OllamaBaseURL:   v.GetString("OLLAMA_BASE_URL"),

		// Social Media
		TwitterAPIKey:        v.GetString("TWITTER_API_KEY"),
		TwitterAPISecret:     v.GetString("TWITTER_API_SECRET"),
		LinkedInClientID:     v.GetString("LINKEDIN_CLIENT_ID"),
		LinkedInClientSecret: v.GetString("LINKEDIN_CLIENT_SECRET"),

		// Email
		SMTPHost:     v.GetString("SMTP_HOST"),
		SMTPPort:     v.GetInt("SMTP_PORT"),
		SMTPUser:     v.GetString("SMTP_USER"),
		SMTPPassword: v.GetString("SMTP_PASSWORD"),

		// Messaging
		SlackClientID:     v.GetString("SLACK_CLIENT_ID"),
		SlackClientSecret: v.GetString("SLACK_CLIENT_SECRET"),
		DiscordBotToken:   v.GetString("DISCORD_BOT_TOKEN"),

		// Monitoring
		SentryDSN: v.GetString("SENTRY_DSN"),
	}

	// Validate required config
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.EncryptionKey == "" && cfg.Environment == "production" {
		return nil, fmt.Errorf("ENCRYPTION_KEY is required in production")
	}

	return cfg, nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

