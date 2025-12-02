package models

import "time"

type ServerConfig struct {
	Host     string             `mapstructure:"host"`
	Port     int                `mapstructure:"port"`
	Limits   ServerLimitsConfig `mapstructure:"limits"`
	Metrics  MetricsConfig      `mapstructure:"metrics"`
	Health   HealthConfig       `mapstructure:"health"`
	Ready    ReadyConfig        `mapstructure:"ready"`
	Security SecurityConfig     `mapstructure:"security"`
}

type ServerLimitsConfig struct {
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	RequestsPerMinute int           `mapstructure:"requests_per_minute"`
	Burst             int           `mapstructure:"burst"`
}

type LoginConfig struct {
	Endpoint string `mapstructure:"endpoint" default:"https://auth.thand.io/"`
	ApiKey   string `mapstructure:"api_key"`          // API key for authenticating with the login server
	Base     string `mapstructure:"base" default:"/"` // Base path for login endpoint e.g. /
}

type LoggingConfig struct {
	Level  string `mapstructure:"level" default:"info"`
	Format string `mapstructure:"format" default:"text"`
	Output string `mapstructure:"output"`
}

type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled" default:"true"`
	Path      string `mapstructure:"path" default:"/metrics"`
	Namespace string `mapstructure:"namespace"`
}

type HealthConfig struct {
	Enabled bool `mapstructure:"enabled" default:"true"`
	// Don't use /healthz as it conflicts with google k8s health checks
	Path string `mapstructure:"path" default:"/health"`
}

type ReadyConfig struct {
	Enabled bool   `mapstructure:"enabled" default:"true"`
	Path    string `mapstructure:"path" default:"/ready"`
}

type SecurityConfig struct {
	CORS CORSConfig `mapstructure:"cors"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// WithDefaults returns a CORSConfig with default values applied for any unset fields
func (c CORSConfig) WithDefaults() CORSConfig {
	if len(c.AllowedMethods) == 0 {
		c.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	}
	if len(c.AllowedHeaders) == 0 {
		c.AllowedHeaders = []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		}
	}
	if c.MaxAge == 0 {
		c.MaxAge = 86400 // 24 hours
	}
	return c
}

// AddOrigins appends additional origins to the allowed list
func (c *CORSConfig) AddOrigins(origins ...string) {
	c.AllowedOrigins = append(c.AllowedOrigins, origins...)
}

type APIConfig struct {
	Version   string          `mapstructure:"version" default:"v1"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

func (api *APIConfig) GetVersion() string {
	if len(api.Version) > 0 {
		return api.Version
	}
	return "v1"
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	Burst             int `mapstructure:"burst"`
}
