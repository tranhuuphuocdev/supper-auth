package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/joho/godotenv"
)

type DomainConfig struct {
	DBName string `json:"db_name"`
	Schema string `json:"schema"`
}

type Config struct {
	Port            string
	JWTSecret       string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPassword      string
	DBSSLMode       string
	DefaultDomain   string
	DefaultDBName   string
	DefaultDBSchema string
	Domains         map[string]DomainConfig
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "postgres"),
		DBPassword:      getEnv("DB_PASSWORD", "password"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		DefaultDomain:   getEnv("DEFAULT_DOMAIN", "default"),
		DefaultDBName:   getEnv("DB_NAME", "auth_service"),
		DefaultDBSchema: getEnv("DB_SCHEMA", "public"),
		Domains:         map[string]DomainConfig{},
	}

	if raw := strings.TrimSpace(os.Getenv("DOMAIN_DB_MAP")); raw != "" {
		if err := json.Unmarshal([]byte(raw), &cfg.Domains); err != nil {
			return nil, fmt.Errorf("invalid DOMAIN_DB_MAP: %w", err)
		}
	}

	if _, ok := cfg.Domains[cfg.DefaultDomain]; !ok {
		cfg.Domains[cfg.DefaultDomain] = DomainConfig{
			DBName: cfg.DefaultDBName,
			Schema: cfg.DefaultDBSchema,
		}
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return nil, fmt.Errorf("JWT_SECRET must not be empty")
	}

	return cfg, nil
}

func (c *Config) ResolveDomain(domain string) (string, DomainConfig) {
	normalized := strings.TrimSpace(strings.ToLower(domain))
	if normalized == "" {
		normalized = c.DefaultDomain
	}

	if dc, ok := c.Domains[normalized]; ok {
		if strings.TrimSpace(dc.DBName) == "" {
			dc.DBName = c.DefaultDBName
		}
		if strings.TrimSpace(dc.Schema) == "" {
			dc.Schema = c.DefaultDBSchema
		}
		return normalized, dc
	}

	return normalized, DomainConfig{
		DBName: c.DefaultDBName,
		Schema: c.DefaultDBSchema,
	}
}

func (c *Config) DomainList() []string {
	domains := make([]string, 0, len(c.Domains))
	for k := range c.Domains {
		domains = append(domains, k)
	}
	sort.Strings(domains)
	return domains
}

func getEnv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
