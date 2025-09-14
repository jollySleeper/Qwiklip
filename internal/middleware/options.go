package middleware

// MiddlewareOption is a function that modifies middleware behavior
type MiddlewareOption func(*MiddlewareConfig)

// MiddlewareConfig holds the configuration for middleware
type MiddlewareConfig struct {
	EnableRecovery bool
	EnableLogging  bool
	EnableCORS     bool
}

// WithRecovery enables error recovery middleware
func WithRecovery() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.EnableRecovery = true
	}
}

// WithLogging enables request logging middleware
func WithLogging() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.EnableLogging = true
	}
}

// WithCORS enables cross-origin resource sharing middleware
func WithCORS() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.EnableCORS = true
	}
}

// DefaultConfig returns a middleware configuration with common defaults
func DefaultConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		EnableRecovery: true,
		EnableLogging:  true,
		EnableCORS:     true,
	}
}

// MinimalConfig returns a middleware configuration with minimal features
func MinimalConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		EnableRecovery: false,
		EnableLogging:  false,
		EnableCORS:     false,
	}
}

// ApplyOptions applies functional options to create middleware configuration
func ApplyOptions(opts ...MiddlewareOption) *MiddlewareConfig {
	config := &MiddlewareConfig{} // Start with all disabled
	for _, opt := range opts {
		opt(config)
	}
	return config
}
