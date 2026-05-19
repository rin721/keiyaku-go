package platform

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr               string
	BaseURL            string
	KeiyakuHost        string
	RegistrationSecret string
	GatewaySecret      string
	InstanceID         string
	MySQLDSN           string
	SnowflakeNode      int64
	HeartbeatInterval  time.Duration
	RegisterTimeout    time.Duration
	ShutdownTimeout    time.Duration
}

func LoadConfig() (Config, error) {
	cfg := Config{
		Addr:               env("BLOG_ADDR", ":9091"),
		BaseURL:            env("BLOG_BASE_URL", ""),
		KeiyakuHost:        env("KEIYAKU_HOST", "http://127.0.0.1:8080"),
		RegistrationSecret: os.Getenv("BLOG_REGISTRATION_SECRET"),
		GatewaySecret:      os.Getenv("BLOG_GATEWAY_SECRET"),
		InstanceID:         env("BLOG_INSTANCE_ID", "blog-local"),
		MySQLDSN:           os.Getenv("BLOG_MYSQL_DSN"),
		SnowflakeNode:      envInt64("BLOG_SNOWFLAKE_NODE", 2),
		HeartbeatInterval:  envDuration("BLOG_HEARTBEAT_INTERVAL", 10*time.Second),
		RegisterTimeout:    envDuration("BLOG_REGISTER_TIMEOUT", 5*time.Second),
		ShutdownTimeout:    envDuration("BLOG_SHUTDOWN_TIMEOUT", 10*time.Second),
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://127.0.0.1" + cfg.Addr
	}
	if strings.TrimSpace(cfg.MySQLDSN) == "" {
		return Config{}, fmt.Errorf("BLOG_MYSQL_DSN is required")
	}
	if strings.TrimSpace(cfg.RegistrationSecret) == "" {
		return Config{}, fmt.Errorf("BLOG_REGISTRATION_SECRET is required")
	}
	if strings.TrimSpace(cfg.GatewaySecret) == "" {
		return Config{}, fmt.Errorf("BLOG_GATEWAY_SECRET is required")
	}
	return cfg, nil
}

func env(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt64(key string, fallback int64) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
