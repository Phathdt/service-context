package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `yaml:"database" mapstructure:"database"`
	Redis    RedisConfig    `yaml:"redis"    mapstructure:"redis"`
	Server   ServerConfig   `yaml:"server"   mapstructure:"server"`
	Worker   WorkerConfig   `yaml:"worker"   mapstructure:"worker"`
	Logger   LoggerConfig   `yaml:"logger"   mapstructure:"logger"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"     mapstructure:"host"`
	Port     int    `yaml:"port"     mapstructure:"port"`
	User     string `yaml:"user"     mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Database string `yaml:"database" mapstructure:"database"`
	SSLMode  string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
}

type RedisConfig struct {
	Host     string `yaml:"host"     mapstructure:"host"`
	Port     int    `yaml:"port"     mapstructure:"port"`
	Password string `yaml:"password" mapstructure:"password"`
	Database int    `yaml:"database" mapstructure:"database"`
}

type ServerConfig struct {
	Host string `yaml:"host" mapstructure:"host"`
	Port int    `yaml:"port" mapstructure:"port"`
}

type WorkerConfig struct {
	Enabled     bool `yaml:"enabled"     mapstructure:"enabled"`
	Concurrency int  `yaml:"concurrency" mapstructure:"concurrency"`
}

type LoggerConfig struct {
	Level  string `yaml:"level"  mapstructure:"level"`
	Format string `yaml:"format" mapstructure:"format"`
}

// GetDatabaseDSN returns a PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

// GetRedisURI returns a Redis connection string
func (c *RedisConfig) GetURI() string {
	if c.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%d/%d", c.Password, c.Host, c.Port, c.Database)
	}
	return fmt.Sprintf("redis://%s:%d/%d", c.Host, c.Port, c.Database)
}

// GetServerAddress returns the server listen address
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// LoadConfig loads configuration from YAML file first, then overrides with environment variables
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Load .env file if exists (optional)
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: failed to load .env file: %v\n", err)
	}

	// Configure viper for YAML file
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Read YAML configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Configure environment variable support
	// This allows env vars like DATABASE__HOST to override database.host
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	v.AutomaticEnv()

	var config Config

	// Unmarshal config with automatic env override
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set defaults if not specified
	if config.Database.SSLMode == "" {
		config.Database.SSLMode = "disable"
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Logger.Level == "" {
		config.Logger.Level = "info"
	}
	if config.Logger.Format == "" {
		config.Logger.Format = "text"
	}

	return &config, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Database: "todoapp",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			Database: 0,
		},
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 4000,
		},
		Worker: WorkerConfig{
			Enabled:     true,
			Concurrency: 10,
		},
		Logger: LoggerConfig{
			Level:  "info",
			Format: "text",
		},
	}
}