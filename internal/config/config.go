package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Email    EmailConfig
}

type AppConfig struct {
	Env  string
	Name string
}

type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
}

type JWTConfig struct {
	Secret          string
	StaffTTL        time.Duration
	TenantTTL       time.Duration
}

type CORSConfig struct {
	AllowedOrigins []string
}

type EmailConfig struct {
	ResendAPIKey     string
	From             string
	FrontendURL      string
	PasswordResetTTL time.Duration
	DevRedirectTo    string // development: send all mail here (Resend sandbox allows only account email)
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Name: getEnv("APP_NAME", "nivas"),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
		},
		Database: DatabaseConfig{
			URL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/nivas?sslmode=disable"),
			MaxConns:        int32(getEnvInt("DB_MAX_CONNS", 25)),
			MinConns:        int32(getEnvInt("DB_MIN_CONNS", 5)),
			MaxConnLifetime: getEnvDuration("DB_MAX_CONN_LIFETIME", time.Hour),
		},
		JWT: JWTConfig{
			Secret:    getEnv("JWT_SECRET", "change-me-in-production"),
			StaffTTL:  getEnvDuration("JWT_STAFF_TTL", 7*24*time.Hour),
			TenantTTL: getEnvDuration("JWT_TENANT_TTL", 30*24*time.Hour),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
		},
		Email: EmailConfig{
			ResendAPIKey:     getEnv("RESEND_API_KEY", ""),
			From:             getEnv("EMAIL_FROM", "Nivas <onboarding@resend.dev>"),
			FrontendURL:      getEnv("FRONTEND_URL", "http://localhost:5173"),
			PasswordResetTTL: getEnvDuration("PASSWORD_RESET_TTL", time.Hour),
			DevRedirectTo:    getEnv("EMAIL_DEV_REDIRECT_TO", ""),
		},
	}

	if cfg.JWT.Secret == "change-me-in-production" && cfg.App.Env == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production")
	}

	return cfg, nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := []string{}
		for _, p := range splitComma(v) {
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) > 0 {
			return parts
		}
	}
	return fallback
}

func splitComma(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return out
}
