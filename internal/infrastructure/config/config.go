package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig       `mapstructure:"app"`
	Server    ServerConfig    `mapstructure:"server"`
	Log       LogConfig       `mapstructure:"log"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Snowflake SnowflakeConfig `mapstructure:"snowflake"`
	Security  SecurityConfig  `mapstructure:"security"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type ServerConfig struct {
	Addr            string        `mapstructure:"addr"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputDir  string `mapstructure:"output_dir"`
	MaxSizeMB  int    `mapstructure:"max_size_mb"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAgeDays int    `mapstructure:"max_age_days"`
	Compress   bool   `mapstructure:"compress"`
}

type MySQLConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	SlowThreshold   time.Duration `mapstructure:"slow_threshold"`
}

type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Username     string        `mapstructure:"username"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	Issuer          string        `mapstructure:"issuer"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

type SnowflakeConfig struct {
	Node int64 `mapstructure:"node"`
}

type SecurityConfig struct {
	BcryptCost int `mapstructure:"bcrypt_cost"`
	RateLimit  struct {
		RequestsPerSecond float64 `mapstructure:"requests_per_second"`
		Burst             int     `mapstructure:"burst"`
	} `mapstructure:"rate_limit"`
	CircuitBreaker struct {
		FailureThreshold uint32        `mapstructure:"failure_threshold"`
		OpenTimeout      time.Duration `mapstructure:"open_timeout"`
	} `mapstructure:"circuit_breaker"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	setDefaults(v)
	v.SetEnvPrefix("KEIYAKU")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && path != "" {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c Config) Validate() error {
	if c.Server.Addr == "" {
		return fmt.Errorf("server.addr is required")
	}
	if c.MySQL.DSN == "" {
		return fmt.Errorf("mysql.dsn is required")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr is required")
	}
	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("jwt.secret must be at least 32 bytes")
	}
	if c.JWT.AccessTokenTTL <= 0 || c.JWT.RefreshTokenTTL <= 0 {
		return fmt.Errorf("jwt ttl must be positive")
	}
	if c.Security.BcryptCost == 0 {
		c.Security.BcryptCost = 12
	}
	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "keiyaku-go")
	v.SetDefault("app.env", "local")
	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.read_timeout", "5s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.shutdown_timeout", "10s")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output_dir", "logs")
	v.SetDefault("log.max_size_mb", 100)
	v.SetDefault("log.max_backups", 7)
	v.SetDefault("log.max_age_days", 30)
	v.SetDefault("log.compress", true)
	v.SetDefault("mysql.max_open_conns", 50)
	v.SetDefault("mysql.max_idle_conns", 10)
	v.SetDefault("mysql.conn_max_lifetime", "1h")
	v.SetDefault("mysql.conn_max_idle_time", "30m")
	v.SetDefault("mysql.slow_threshold", "200ms")
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.dial_timeout", "3s")
	v.SetDefault("redis.read_timeout", "1s")
	v.SetDefault("redis.write_timeout", "1s")
	v.SetDefault("jwt.issuer", "keiyaku-go")
	v.SetDefault("jwt.access_token_ttl", "15m")
	v.SetDefault("jwt.refresh_token_ttl", "168h")
	v.SetDefault("snowflake.node", 1)
	v.SetDefault("security.bcrypt_cost", 12)
	v.SetDefault("security.rate_limit.requests_per_second", 100)
	v.SetDefault("security.rate_limit.burst", 200)
	v.SetDefault("security.circuit_breaker.failure_threshold", 5)
	v.SetDefault("security.circuit_breaker.open_timeout", "5s")
}
