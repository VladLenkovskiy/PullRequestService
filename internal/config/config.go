package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Logger   LoggerConfig
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type ServerConfig struct {
	Port string
}

type LoggerConfig struct {
	Level string
}

func (c DatabaseConfig) ConnString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found, using environment variables")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:            GetEnv("DB_HOST", "localhost"),
			Port:            GetEnvAsInt("DB_PORT", 5432),
			User:            GetEnv("DB_USER", "postgres"),
			Password:        GetEnv("DB_PASSWORD", "postgres"),
			DBName:          GetEnv("DB_NAME", "pr_service"),
			SSLMode:         GetEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:    GetEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    GetEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: GetEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Server: ServerConfig{
			Port: GetEnv("SERVER_PORT", "8080"),
		},
		Logger: LoggerConfig{
			Level: GetEnv("LOG_LEVEL", "info"),
		},
	}, nil
}

func GetEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func GetEnvAsInt(key string, defaultVal int) int {
	valueStr := GetEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	value, _ := strconv.Atoi(valueStr)
	return value
}

func GetEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := GetEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	value, _ := time.ParseDuration(valueStr)
	return value
}
