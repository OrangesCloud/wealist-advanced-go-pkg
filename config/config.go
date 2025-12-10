// Package config provides common configuration loading for all Go services.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for a service
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	Services ServicesConfig `yaml:"services"`
	CORS     CORSConfig     `yaml:"cors"`
	S3       S3Config       `yaml:"s3"`
	Logger   LoggerConfig   `yaml:"logger"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port            int           `yaml:"port"`
	Mode            string        `yaml:"mode"` // debug, release
	BasePath        string        `yaml:"base_path"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            string        `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	SSLMode         string        `yaml:"sslmode"`
	URL             string        `yaml:"url"` // DATABASE_URL format
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	URL      string `yaml:"url"` // redis:// format
	TLS      bool   `yaml:"tls"`
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret     string        `yaml:"secret"`
	ExpireTime time.Duration `yaml:"expire_time"`
}

// ServicesConfig holds external service URLs
type ServicesConfig struct {
	AuthServiceURL    string        `yaml:"auth_service_url"`
	UserServiceURL    string        `yaml:"user_service_url"`
	BoardServiceURL   string        `yaml:"board_service_url"`
	ChatServiceURL    string        `yaml:"chat_service_url"`
	NotiServiceURL    string        `yaml:"noti_service_url"`
	StorageServiceURL string        `yaml:"storage_service_url"`
	VideoServiceURL   string        `yaml:"video_service_url"`
	Timeout           time.Duration `yaml:"timeout"`
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins string `yaml:"allowed_origins"`
}

// S3Config holds S3/MinIO configuration
type S3Config struct {
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Endpoint  string `yaml:"endpoint"` // For MinIO
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level      string `yaml:"level"` // debug, info, warn, error
	OutputPath string `yaml:"output_path"`
}

// DefaultConfig returns default configuration values
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            8080,
			Mode:            "debug",
			ReadTimeout:     10 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 30 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            "5432",
			User:            "postgres",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		JWT: JWTConfig{
			ExpireTime: 24 * time.Hour,
		},
		Services: ServicesConfig{
			Timeout: 5 * time.Second,
		},
		CORS: CORSConfig{
			AllowedOrigins: "*",
		},
		Logger: LoggerConfig{
			Level:      "info",
			OutputPath: "stdout",
		},
	}
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Try to read config file (optional)
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
		}
	}

	// Override with environment variables
	cfg.LoadFromEnv()

	return cfg, nil
}

// LoadFromEnv overrides configuration with environment variables
func (c *Config) LoadFromEnv() {
	// Server
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		c.Server.Mode = mode
	}
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		c.Server.Mode = mode
	}
	// ENV alias: dev→debug, prod→release
	if env := os.Getenv("ENV"); env != "" {
		switch env {
		case "dev":
			c.Server.Mode = "debug"
		case "prod":
			c.Server.Mode = "release"
		default:
			c.Server.Mode = env
		}
	}
	if basePath := os.Getenv("SERVER_BASE_PATH"); basePath != "" {
		c.Server.BasePath = basePath
	}

	// Database - DATABASE_URL takes precedence
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		c.Database.URL = dbURL
		c.parseDatabaseURL(dbURL)
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		c.Database.Port = port
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		c.Database.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		c.Database.DBName = dbname
	}

	// Redis
	if host := os.Getenv("REDIS_HOST"); host != "" {
		c.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Redis.Port = p
		}
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		c.Redis.Password = password
	}
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		c.Redis.URL = redisURL
	}

	// JWT
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		c.JWT.Secret = secret
	}
	if secret := os.Getenv("SECRET_KEY"); secret != "" {
		c.JWT.Secret = secret
	}

	// Services
	if url := os.Getenv("AUTH_SERVICE_URL"); url != "" {
		c.Services.AuthServiceURL = url
	}
	if url := os.Getenv("USER_SERVICE_URL"); url != "" {
		c.Services.UserServiceURL = url
	}
	if url := os.Getenv("BOARD_SERVICE_URL"); url != "" {
		c.Services.BoardServiceURL = url
	}
	if url := os.Getenv("CHAT_SERVICE_URL"); url != "" {
		c.Services.ChatServiceURL = url
	}
	if url := os.Getenv("NOTI_SERVICE_URL"); url != "" {
		c.Services.NotiServiceURL = url
	}
	if url := os.Getenv("STORAGE_SERVICE_URL"); url != "" {
		c.Services.StorageServiceURL = url
	}
	if url := os.Getenv("VIDEO_SERVICE_URL"); url != "" {
		c.Services.VideoServiceURL = url
	}

	// CORS
	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		c.CORS.AllowedOrigins = origins
	}
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		c.CORS.AllowedOrigins = origins
	}

	// S3
	if bucket := os.Getenv("S3_BUCKET"); bucket != "" {
		c.S3.Bucket = bucket
	}
	if region := os.Getenv("S3_REGION"); region != "" {
		c.S3.Region = region
	}
	if accessKey := os.Getenv("S3_ACCESS_KEY"); accessKey != "" {
		c.S3.AccessKey = accessKey
	}
	if secretKey := os.Getenv("S3_SECRET_KEY"); secretKey != "" {
		c.S3.SecretKey = secretKey
	}
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		c.S3.Endpoint = endpoint
	}

	// Logger
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Logger.Level = level
	}
}

// parseDatabaseURL parses DATABASE_URL and populates individual fields
func (c *Config) parseDatabaseURL(databaseURL string) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return
	}

	if u.User != nil {
		c.Database.User = u.User.Username()
		c.Database.Password, _ = u.User.Password()
	}

	if u.Host != "" {
		hostPort := u.Host
		if strings.Contains(hostPort, ":") {
			parts := strings.Split(hostPort, ":")
			c.Database.Host = parts[0]
			c.Database.Port = parts[1]
		} else {
			c.Database.Host = hostPort
			c.Database.Port = "5432"
		}
	}

	c.Database.DBName = strings.TrimPrefix(u.Path, "/")
	if sslmode := u.Query().Get("sslmode"); sslmode != "" {
		c.Database.SSLMode = sslmode
	}
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	if c.URL != "" {
		return c.URL
	}
	sslmode := c.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, sslmode,
	)
}

// GetRedisAddr returns Redis address in host:port format
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
