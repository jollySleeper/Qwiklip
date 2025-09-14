package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server    ServerConfig
	Instagram InstagramConfig
	Logging   LoggingConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// InstagramConfig holds Instagram client configuration
type InstagramConfig struct {
	Timeout   time.Duration
	UserAgent string
	Debug     bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// Load loads configuration from environment variables with sensible defaults
func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 300 * time.Second, // Longer for video streaming
			IdleTimeout:  120 * time.Second,
		},
		Instagram: InstagramConfig{
			Timeout:   30 * time.Second,
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			Debug:     getEnvAsBool("DEBUG", false),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "text"), // text or json
		},
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// Validate performs comprehensive validation of all configuration values
func (c *Config) Validate() error {
	if err := c.validateServerConfig(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}

	if err := c.validateInstagramConfig(); err != nil {
		return fmt.Errorf("instagram config: %w", err)
	}

	if err := c.validateLoggingConfig(); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	return nil
}

// validateServerConfig validates server-related configuration
func (c *Config) validateServerConfig() error {
	// Validate port
	portNum, err := strconv.Atoi(c.Server.Port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", c.Server.Port)
	}
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port number out of range (1-65535): %d", portNum)
	}

	// Validate timeouts
	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive, got %v", c.Server.ReadTimeout)
	}
	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive, got %v", c.Server.WriteTimeout)
	}
	if c.Server.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive, got %v", c.Server.IdleTimeout)
	}

	// Read timeout should be reasonable (not too long for security)
	if c.Server.ReadTimeout > 5*time.Minute {
		return fmt.Errorf("read timeout too long (max 5m), got %v", c.Server.ReadTimeout)
	}

	// Write timeout should be reasonable for video streaming
	if c.Server.WriteTimeout < 30*time.Second {
		return fmt.Errorf("write timeout too short for video streaming (min 30s), got %v", c.Server.WriteTimeout)
	}

	return nil
}

// validateInstagramConfig validates Instagram client configuration
func (c *Config) validateInstagramConfig() error {
	// Validate timeout
	if c.Instagram.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", c.Instagram.Timeout)
	}
	if c.Instagram.Timeout > 5*time.Minute {
		return fmt.Errorf("timeout too long (max 5m), got %v", c.Instagram.Timeout)
	}

	// Validate user agent
	if strings.TrimSpace(c.Instagram.UserAgent) == "" {
		return fmt.Errorf("user agent cannot be empty")
	}
	if len(c.Instagram.UserAgent) < 10 {
		return fmt.Errorf("user agent too short (min 10 chars), got %d", len(c.Instagram.UserAgent))
	}
	if len(c.Instagram.UserAgent) > 500 {
		return fmt.Errorf("user agent too long (max 500 chars), got %d", len(c.Instagram.UserAgent))
	}

	return nil
}

// validateLoggingConfig validates logging configuration
func (c *Config) validateLoggingConfig() error {
	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[strings.ToLower(c.Logging.Level)] {
		return fmt.Errorf("invalid log level '%s', must be one of: debug, info, warn, error", c.Logging.Level)
	}

	// Validate log format
	validFormats := map[string]bool{
		"text": true,
		"json": true,
	}
	if !validFormats[strings.ToLower(c.Logging.Format)] {
		return fmt.Errorf("invalid log format '%s', must be one of: text, json", c.Logging.Format)
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}
