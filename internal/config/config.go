package config

import "os"

type ServerConfig struct{
  Port string
  DBAddress string
  DBUser string
  DBPassword string
  DBName string
}

func LoadServerConfig() *ServerConfig{
  return &ServerConfig{
    Port:       getEnv("SERVER_PORT", "9090"),
    DBAddress:  getEnv("DB_ADDRESS", "localhost:9000"),
    DBUser:     getEnv("DB_USER", "user"),
    DBPassword: getEnv("DB_PASSWORD", "password"),
    DBName:     getEnv("DB_NAME", "audit"),
  }
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
