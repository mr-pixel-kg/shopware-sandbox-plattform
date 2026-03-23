package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Logging  LoggingConfig
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Sandbox  SandboxConfig
	Docker   DockerConfig
	Storage  StorageConfig
	Guard    GuardConfig
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type LogFormat string

const (
	LogFormatJSON LogFormat = "json"
	LogFormatText LogFormat = "text"
)

type LoggingConfig struct {
	Level  LogLevel
	Format LogFormat
}

type ServerConfig struct {
	Port           int
	BaseURL        string
	AllowedOrigins []string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

type AuthConfig struct {
	JWTSecret          string
	JWTTTLMinutes      int
	GuestJWTTTLMinutes int
	GuestCookieName    string
}

type SandboxConfig struct {
	HostSuffix      string
	URLPrefix       string
	DefaultTTL      time.Duration
	MaxTTL          time.Duration
	CleanupInterval time.Duration
	InternalPort    int
}

type DockerMode string

const (
	DockerModePort    DockerMode = "port"
	DockerModeTraefik DockerMode = "traefik"
)

type DockerConfig struct {
	Mode                DockerMode
	Network             string
	TrustedProxies      string
	TraefikEnable       bool
	TraefikEntrypoints  string
	TraefikCertResolver string
	TraefikMiddlewares  string
	SnapshotAuthor      string
	SnapshotComment     string
}

type StorageConfig struct {
	ThumbnailDir string
}

type GuardConfig struct {
	MaxActiveTotal      int
	MaxPublicDemosPerIP int
	MaxActivePerUser    int
}

func MustLoad() Config {
	cfgPath := getEnv("CONFIG_PATH", "config.yml")

	v := viper.New()
	v.SetConfigFile(cfgPath)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("read config file %s: %v", cfgPath, err))
	}

	// Map the raw Viper values into typed config structs once so the rest of the
	// application can avoid stringly-typed lookups.
	cfg := Config{
		Logging: LoggingConfig{
			Level:  LogLevel(v.GetString("logging.level")),
			Format: LogFormat(v.GetString("logging.format")),
		},
		Server: ServerConfig{
			Port:           v.GetInt("server.port"),
			BaseURL:        v.GetString("server.app_url"),
			AllowedOrigins: v.GetStringSlice("server.allowed_origins"),
		},
		Database: DatabaseConfig{
			Host:     v.GetString("database.host"),
			Port:     v.GetInt("database.port"),
			Name:     v.GetString("database.name"),
			User:     v.GetString("database.user"),
			Password: v.GetString("database.password"),
			SSLMode:  v.GetString("database.sslmode"),
		},
		Auth: AuthConfig{
			JWTSecret:          v.GetString("auth.jwt_secret"),
			JWTTTLMinutes:      v.GetInt("auth.jwt_ttl_minutes"),
			GuestJWTTTLMinutes: v.GetInt("auth.guest_jwt_ttl_minutes"),
			GuestCookieName:    v.GetString("auth.guest_cookie_name"),
		},
		Sandbox: SandboxConfig{
			HostSuffix:      v.GetString("sandbox.url_suffix"),
			URLPrefix:       v.GetString("sandbox.url_prefix"),
			DefaultTTL:      time.Duration(v.GetInt("sandbox.default_lifetime")) * time.Minute,
			MaxTTL:          time.Duration(v.GetInt("sandbox.max_lifetime")) * time.Minute,
			CleanupInterval: time.Duration(v.GetInt("sandbox.cleanup_interval_seconds")) * time.Second,
			InternalPort:    v.GetInt("sandbox.internal_port"),
		},
		Docker: DockerConfig{
			Mode:                DockerMode(v.GetString("docker.mode")),
			Network:             v.GetString("docker.network"),
			TrustedProxies:      v.GetString("docker.trusted_proxies"),
			TraefikEnable:       v.GetBool("docker.traefik_enable"),
			TraefikEntrypoints:  v.GetString("docker.traefik_entrypoints"),
			TraefikCertResolver: v.GetString("docker.traefik_certresolver"),
			TraefikMiddlewares:  v.GetString("docker.traefik_middlewares"),
			SnapshotAuthor:      v.GetString("docker.snapshot_author"),
			SnapshotComment:     v.GetString("docker.snapshot_comment"),
		},
		Storage: StorageConfig{
			ThumbnailDir: v.GetString("storage.thumbnail_dir"),
		},
		Guard: GuardConfig{
			MaxActiveTotal:      v.GetInt("guard.max_total_sandboxes"),
			MaxPublicDemosPerIP: v.GetInt("guard.max_sandboxes_per_ip"),
			MaxActivePerUser:    v.GetInt("guard.max_sandboxes_per_user"),
		},
	}

	if cfg.Logging.Level == "" {
		cfg.Logging.Level = LogLevelInfo
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = LogFormatJSON
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Auth.JWTSecret == "" {
		panic("missing auth.jwt_secret in config")
	}
	if cfg.Auth.JWTTTLMinutes == 0 {
		cfg.Auth.JWTTTLMinutes = 480
	}
	if cfg.Auth.GuestJWTTTLMinutes == 0 {
		cfg.Auth.GuestJWTTTLMinutes = 43200
	}
	if cfg.Auth.GuestCookieName == "" {
		cfg.Auth.GuestCookieName = "shopshredder_guest"
	}
	if cfg.Sandbox.DefaultTTL == 0 {
		cfg.Sandbox.DefaultTTL = 60 * time.Minute
	}
	if cfg.Sandbox.MaxTTL == 0 {
		cfg.Sandbox.MaxTTL = 24 * time.Hour
	}
	if cfg.Sandbox.CleanupInterval == 0 {
		cfg.Sandbox.CleanupInterval = 60 * time.Second
	}
	if cfg.Sandbox.InternalPort == 0 {
		cfg.Sandbox.InternalPort = 80
	}
	if cfg.Docker.Mode == "" {
		cfg.Docker.Mode = DockerModePort
	}
	if cfg.Docker.Mode != DockerModePort && cfg.Docker.Mode != DockerModeTraefik {
		panic(fmt.Sprintf("invalid docker.mode %q: must be \"port\" or \"traefik\"", cfg.Docker.Mode))
	}
	if cfg.Docker.Network == "" {
		cfg.Docker.Network = "internal"
	}
	if cfg.Docker.TrustedProxies == "" {
		cfg.Docker.TrustedProxies = "0.0.0.0/0"
	}
	if cfg.Docker.TraefikEntrypoints == "" {
		cfg.Docker.TraefikEntrypoints = "websecure"
	}
	if cfg.Docker.TraefikCertResolver == "" {
		cfg.Docker.TraefikCertResolver = "production"
	}
	if cfg.Docker.TraefikMiddlewares == "" {
		cfg.Docker.TraefikMiddlewares = "sandbox-middleware@file,https-redirect@file"
	}
	if cfg.Docker.SnapshotAuthor == "" {
		cfg.Docker.SnapshotAuthor = "shopshredder-api"
	}
	if cfg.Docker.SnapshotComment == "" {
		cfg.Docker.SnapshotComment = "Sandbox snapshot created by Shopshredder API"
	}
	if cfg.Storage.ThumbnailDir == "" {
		cfg.Storage.ThumbnailDir = "storage/thumbnails"
	}

	// Defaults are applied after reading YAML so partially filled config files
	// stay valid in local development and in container environments.
	return cfg
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host,
		c.Port,
		c.Name,
		c.User,
		c.Password,
		c.SSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
