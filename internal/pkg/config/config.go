package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env     string
	Server  ServerConfig
	DB      DBConfig
	Redis   RedisConfig
	NATS    NATSConfig
	JWT     JWTConfig
	Session SessionConfig
	Zone    ZoneConfig
	Logging LoggingConfig
	Metrics MetricsConfig
}

type ServerConfig struct {
	APIPort  int
	ZonePort int
}

type DBConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type NATSConfig struct {
	URL string
}

type JWTConfig struct{
	Secret                 string
	PrivateKeyPath         string
	PublicKeyPath          string
	AccessTokenDuration    time.Duration
	RefreshTokenDuration   time.Duration
}

type SessionConfig struct {
	Duration                time.Duration
	ConnectionTokenDuration time.Duration
	TransferTokenDuration   time.Duration
}

type ZoneConfig struct {
	ID                  string
	Shard               string
	MaxPlayers          int
	TickRateMs          int
	HeartbeatIntervalMs int
}

type LoggingConfig struct {
	Level  string
	Format string
}

type MetricsConfig struct {
	Enabled bool
	Port    int
}

// Load loads configuration from environment variables and config files
func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Try to read config file, ignore error if not found
	_ = viper.ReadInConfig()

	config := &Config{
		Env: viper.GetString("ENV"),
		Server: ServerConfig{
			APIPort:  viper.GetInt("API_PORT"),
			ZonePort: viper.GetInt("ZONE_PORT"),
		},
		DB: DBConfig{
			Host:         viper.GetString("DB_HOST"),
			Port:         viper.GetInt("DB_PORT"),
			User:         viper.GetString("DB_USER"),
			Password:     viper.GetString("DB_PASSWORD"),
			Name:         viper.GetString("DB_NAME"),
			SSLMode:      viper.GetString("DB_SSL_MODE"),
			MaxOpenConns: viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: viper.GetInt("DB_MAX_IDLE_CONNS"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("REDIS_HOST"),
			Port:     viper.GetInt("REDIS_PORT"),
			Password: viper.GetString("REDIS_PASSWORD"),
			DB:       viper.GetInt("REDIS_DB"),
		},
		NATS: NATSConfig{
			URL: viper.GetString("NATS_URL"),
		},
		JWT: JWTConfig{
			Secret:                 viper.GetString("JWT_SECRET"),
			PrivateKeyPath:         viper.GetString("JWT_PRIVATE_KEY_PATH"),
			PublicKeyPath:          viper.GetString("JWT_PUBLIC_KEY_PATH"),
			AccessTokenDuration:    viper.GetDuration("JWT_ACCESS_TOKEN_DURATION"),
			RefreshTokenDuration:   viper.GetDuration("JWT_REFRESH_TOKEN_DURATION"),
		},
		Session: SessionConfig{
			Duration:                viper.GetDuration("SESSION_DURATION"),
			ConnectionTokenDuration: viper.GetDuration("CONNECTION_TOKEN_DURATION"),
			TransferTokenDuration:   viper.GetDuration("TRANSFER_TOKEN_DURATION"),
		},
		Zone: ZoneConfig{
			ID:                  viper.GetString("ZONE_ID"),
			Shard:               viper.GetString("ZONE_SHARD"),
			MaxPlayers:          viper.GetInt("ZONE_MAX_PLAYERS"),
			TickRateMs:          viper.GetInt("ZONE_TICK_RATE_MS"),
			HeartbeatIntervalMs: viper.GetInt("ZONE_HEARTBEAT_INTERVAL_MS"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("LOG_LEVEL"),
			Format: viper.GetString("LOG_FORMAT"),
		},
		Metrics: MetricsConfig{
			Enabled: viper.GetBool("METRICS_ENABLED"),
			Port:    viper.GetInt("METRICS_PORT"),
		},
	}

	return config, nil
}

func setDefaults() {
	// Server
	viper.SetDefault("ENV", "development")
	viper.SetDefault("API_PORT", 8080)
	viper.SetDefault("ZONE_PORT", 8081)

	// Database
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "hpmmo")
	viper.SetDefault("DB_PASSWORD", "hpmmo_dev_password")
	viper.SetDefault("DB_NAME", "hpmmo")
	viper.SetDefault("DB_SSL_MODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 5)

	// Redis
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("REDIS_PASSWORD", "hpmmo_redis_password")
	viper.SetDefault("REDIS_DB", 0)

	// NATS
	viper.SetDefault("NATS_URL", "nats://localhost:4222")

	// JWT
	viper.SetDefault("JWT_SECRET", "change_this_in_production")
	viper.SetDefault("JWT_ACCESS_TOKEN_DURATION", 24*time.Hour)
	viper.SetDefault("JWT_REFRESH_TOKEN_DURATION", 30*24*time.Hour)

	// Session
	viper.SetDefault("SESSION_DURATION", 24*time.Hour)
	viper.SetDefault("CONNECTION_TOKEN_DURATION", 30*time.Second)
	viper.SetDefault("TRANSFER_TOKEN_DURATION", 30*time.Second)

	// Zone
	viper.SetDefault("ZONE_ID", "hogwarts_main")
	viper.SetDefault("ZONE_SHARD", "shard-0")
	viper.SetDefault("ZONE_MAX_PLAYERS", 500)
	viper.SetDefault("ZONE_TICK_RATE_MS", 50)
	viper.SetDefault("ZONE_HEARTBEAT_INTERVAL_MS", 10000)

	// Logging
	viper.SetDefault("LOG_LEVEL", "debug")
	viper.SetDefault("LOG_FORMAT", "json")

	// Metrics
	viper.SetDefault("METRICS_ENABLED", true)
	viper.SetDefault("METRICS_PORT", 9091)
}

func (c *Config) GetDBConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.Name, c.DB.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}
