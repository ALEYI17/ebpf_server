package config

import (
	"os"
	"strconv"
)

type ServerConfig struct{
  Port string
  DBAddress string
  DBUser string
  DBPassword string
  DBName string
  BatchSize int
  BatchFlushMs int
  PrometheusPort string
}

func LoadServerConfig() *ServerConfig{
  return &ServerConfig{
    Port:       getEnv("SERVER_PORT", "8080"),
    DBAddress:  getEnv("DB_ADDRESS", "localhost:9000"),
    DBUser:     getEnv("DB_USER", "user"),
    DBPassword: getEnv("DB_PASSWORD", "password"),
    DBName:     getEnv("DB_NAME", "audit"),
    BatchSize: getEnvAsInt("BATCH_MAX_SIZE", 1000),
    BatchFlushMs: getEnvAsInt("BATCH_FLUSH_MS", 10000),
    PrometheusPort: getEnv("PROMETHEUS_PORT", ":9090"),
  }
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}

func getEnvAsInt(name string, defaultVal int) int {
  if valStr := os.Getenv(name); valStr != "" {
    if val, err := strconv.Atoi(valStr); err == nil {
      return val
    }
  }
  return defaultVal
}
