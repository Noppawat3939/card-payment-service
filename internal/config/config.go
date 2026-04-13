package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// relate by env
type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	RedisHost string
	RedisPort string
}

func Load() *Config {
	// get env from .local
	e := godotenv.Load("env.local")
	if e != nil {
		log.Println("warning: .env.local not found, using system env")
	}

	return &Config{
		AppPort:    getEnv("APP_PORT"),
		DBHost:     getEnv("DB_HOST"),
		DBPort:     getEnv("DB_PORT"),
		DBUser:     getEnv("DB_USER"),
		DBPassword: getEnv("DB_PASSWORD"),
		DBName:     getEnv("DB_NAME"),
		RedisHost:  getEnv("REDIS_HOST"),
		RedisPort:  getEnv("REDIS_PORT"),
	}
}

func (c *Config) GetPostgreslDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf(
		"%s:%s",
		c.RedisHost,
		c.RedisPort,
	)
}

func getEnv(key string) string {
	return os.Getenv(key)
}
