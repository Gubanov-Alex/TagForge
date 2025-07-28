package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `envconfig:"SERVER"`
	Database DatabaseConfig `envconfig:"DATABASE"`
	Redis    RedisConfig    `envconfig:"REDIS"`
	Kafka    KafkaConfig    `envconfig:"KAFKA"`
	Logger   LoggerConfig   `envconfig:"LOGGER"`
	Metrics  MetricsConfig  `envconfig:"METRICS"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host         string        `envconfig:"HOST" default:"0.0.0.0"`
	Port         string        `envconfig:"PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"READ_TIMEOUT" default:"30s"`
	WriteTimeout time.Duration `envconfig:"WRITE_TIMEOUT" default:"30s"`
	IdleTimeout  time.Duration `envconfig:"IDLE_TIMEOUT" default:"120s"`
	Environment  string        `envconfig:"ENVIRONMENT" default:"development"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Host            string        `envconfig:"HOST" default:"localhost"`
	Port            string        `envconfig:"PORT" default:"5432"`
	User            string        `envconfig:"USER" default:"postgres"`
	Password        string        `envconfig:"PASSWORD" default:"postgres"`
	Name            string        `envconfig:"NAME" default:"config_service"`
	SSLMode         string        `envconfig:"SSL_MODE" default:"disable"`
	MaxOpenConns    int           `envconfig:"MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `envconfig:"MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetime time.Duration `envconfig:"CONN_MAX_LIFETIME" default:"5m"`
	MigrationsPath  string        `envconfig:"MIGRATIONS_PATH" default:"file://migrations"`
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     string `envconfig:"PORT" default:"6379"`
	Password string `envconfig:"PASSWORD" default:""`
	DB       int    `envconfig:"DB" default:"0"`
}

// KafkaConfig contains Kafka connection configuration
type KafkaConfig struct {
	Brokers []string `envconfig:"BROKERS" default:"localhost:9092"`
	Topic   string   `envconfig:"TOPIC" default:"config-events"`
}

// LoggerConfig contains logging configuration
type LoggerConfig struct {
	Level  string `envconfig:"LEVEL" default:"info"`
	Format string `envconfig:"FORMAT" default:"json"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	Enabled bool   `envconfig:"ENABLED" default:"true"`
	Path    string `envconfig:"PATH" default:"/metrics"`
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetDSN returns PostgreSQL connection string
func (d DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

// GetRedisAddr returns Redis connection address
func (r RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}
