package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/text/language"
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
	Plugins   PluginsConfig   `mapstructure:"plugins"`
	I18N      I18NConfig      `mapstructure:"i18n"`
	RBAC      RBACConfig      `mapstructure:"rbac"`
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

type I18NConfig struct {
	Default   string            `mapstructure:"default"`
	Supported []string          `mapstructure:"supported"`
	Files     map[string]string `mapstructure:"files"`
}

type RBACConfig struct {
	ModelPath  string `mapstructure:"model_path"`
	PolicyPath string `mapstructure:"policy_path"`
}

type PluginsConfig struct {
	Enabled              bool          `mapstructure:"enabled"`
	RegistrationTokens   []string      `mapstructure:"registration_tokens"`
	AllowedPluginKeys    []string      `mapstructure:"allowed_plugin_keys"`
	PublicPrefix         string        `mapstructure:"public_prefix"`
	HeartbeatTTL         time.Duration `mapstructure:"heartbeat_ttl"`
	RequestTimeout       time.Duration `mapstructure:"request_timeout"`
	AllowedHosts         []string      `mapstructure:"allowed_hosts"`
	AllowedCIDRs         []string      `mapstructure:"allowed_cidrs"`
	AllowLoopback        bool          `mapstructure:"allow_loopback"`
	AllowPublicRoutes    bool          `mapstructure:"allow_public_routes"`
	GatewaySigningSecret string        `mapstructure:"gateway_signing_secret"`
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
		if path != "" {
			return nil, fmt.Errorf("read config: %w", err)
		}
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	if err := cfg.normalizeI18N(); err != nil {
		return nil, err
	}
	cfg.resolvePaths(configBaseDir(v.ConfigFileUsed()))
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c == nil {
		return fmt.Errorf("config is required")
	}
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
	if err := c.I18N.Validate(); err != nil {
		return err
	}
	if err := c.RBAC.Validate(); err != nil {
		return err
	}
	if err := c.Plugins.Validate(c.App.Env); err != nil {
		return err
	}
	return nil
}

func (c I18NConfig) Validate() error {
	if c.Default == "" {
		return fmt.Errorf("i18n.default is required")
	}
	if _, err := language.Parse(c.Default); err != nil {
		return fmt.Errorf("parse i18n.default %q: %w", c.Default, err)
	}
	if len(c.Supported) == 0 {
		return fmt.Errorf("i18n.supported is required")
	}
	if len(c.Files) == 0 {
		return fmt.Errorf("i18n.files is required")
	}
	supported := make(map[string]struct{}, len(c.Supported))
	for _, tag := range c.Supported {
		if strings.TrimSpace(tag) == "" {
			return fmt.Errorf("i18n.supported contains empty language")
		}
		parsed, err := language.Parse(tag)
		if err != nil {
			return fmt.Errorf("parse i18n.supported %q: %w", tag, err)
		}
		canonical := parsed.String()
		supported[canonical] = struct{}{}
		if strings.TrimSpace(c.Files[canonical]) == "" {
			return fmt.Errorf("i18n.files.%s is required", tag)
		}
	}
	defaultTag, err := language.Parse(c.Default)
	if err != nil {
		return fmt.Errorf("parse i18n.default %q: %w", c.Default, err)
	}
	if _, ok := supported[defaultTag.String()]; !ok {
		return fmt.Errorf("i18n.default must be included in i18n.supported")
	}
	return nil
}

func (c RBACConfig) Validate() error {
	if c.ModelPath == "" {
		return fmt.Errorf("rbac.model_path is required")
	}
	if c.PolicyPath == "" {
		return fmt.Errorf("rbac.policy_path is required")
	}
	return nil
}

func (c PluginsConfig) Validate(env string) error {
	if !c.Enabled {
		return nil
	}
	if c.PublicPrefix == "" {
		return fmt.Errorf("plugins.public_prefix is required")
	}
	if !strings.HasPrefix(c.PublicPrefix, "/") {
		return fmt.Errorf("plugins.public_prefix must start with /")
	}
	if c.HeartbeatTTL <= 0 {
		return fmt.Errorf("plugins.heartbeat_ttl must be positive")
	}
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("plugins.request_timeout must be positive")
	}
	if env != "" && env != "local" && env != "test" {
		if len(c.RegistrationTokens) == 0 {
			return fmt.Errorf("plugins.registration_tokens is required outside local/test")
		}
		for _, token := range c.RegistrationTokens {
			if len(token) < 32 {
				return fmt.Errorf("plugins.registration_tokens entries must be at least 32 bytes outside local/test")
			}
		}
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
	v.SetDefault("plugins.enabled", true)
	v.SetDefault("plugins.registration_tokens", []string{})
	v.SetDefault("plugins.allowed_plugin_keys", []string{})
	v.SetDefault("plugins.public_prefix", "/api/v1/extensions")
	v.SetDefault("plugins.heartbeat_ttl", "30s")
	v.SetDefault("plugins.request_timeout", "5s")
	v.SetDefault("plugins.allowed_hosts", []string{})
	v.SetDefault("plugins.allowed_cidrs", []string{})
	v.SetDefault("plugins.allow_loopback", false)
	v.SetDefault("plugins.allow_public_routes", false)
	v.SetDefault("plugins.gateway_signing_secret", "")
	v.SetDefault("i18n.default", "en-US")
	v.SetDefault("i18n.supported", []string{"en-US", "zh-CN"})
	v.SetDefault("i18n.files", map[string]string{
		"en-US": "i18n/en-US.yaml",
		"zh-CN": "i18n/zh-CN.yaml",
	})
	v.SetDefault("rbac.model_path", "rbac/model.conf")
	v.SetDefault("rbac.policy_path", "rbac/policy.csv")
}

func (c *Config) resolvePaths(baseDir string) {
	if c == nil {
		return
	}
	for tag, path := range c.I18N.Files {
		c.I18N.Files[tag] = resolvePath(baseDir, path)
	}
	c.RBAC.ModelPath = resolvePath(baseDir, c.RBAC.ModelPath)
	c.RBAC.PolicyPath = resolvePath(baseDir, c.RBAC.PolicyPath)
}

func (c *Config) normalizeI18N() error {
	if c == nil {
		return nil
	}
	if c.I18N.Default != "" {
		tag, err := language.Parse(c.I18N.Default)
		if err != nil {
			return fmt.Errorf("parse i18n.default %q: %w", c.I18N.Default, err)
		}
		c.I18N.Default = tag.String()
	}
	for index, raw := range c.I18N.Supported {
		tag, err := language.Parse(raw)
		if err != nil {
			return fmt.Errorf("parse i18n.supported %q: %w", raw, err)
		}
		c.I18N.Supported[index] = tag.String()
	}
	files := make(map[string]string, len(c.I18N.Files))
	for raw, path := range c.I18N.Files {
		tag, err := language.Parse(raw)
		if err != nil {
			return fmt.Errorf("parse i18n.files language %q: %w", raw, err)
		}
		files[tag.String()] = path
	}
	c.I18N.Files = files
	return nil
}

func configBaseDir(configFile string) string {
	if configFile == "" {
		return "configs"
	}
	return filepath.Dir(configFile)
}

func resolvePath(baseDir string, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	if baseDir == "" {
		baseDir = "."
	}
	abs, err := filepath.Abs(filepath.Join(baseDir, path))
	if err != nil {
		return filepath.Clean(filepath.Join(baseDir, path))
	}
	return abs
}
