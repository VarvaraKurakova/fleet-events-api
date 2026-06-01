package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTP         HTTPConfig
	Postgres     PostgresConfig
	Redis        RedisConfig
	RabbitMQ     RabbitMQConfig
	Worker       WorkerConfig
	DeviceAPIKey string
}

type HTTPConfig struct {
	Addr string
}

type PostgresConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       string
}

type RabbitMQConfig struct {
	URL string
}

type WorkerConfig struct {
	Concurrency int
	Prefetch    int
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr: getEnv("HTTP_ADDR", ":8080"),
		},
		Postgres: PostgresConfig{
			DSN: getEnv("POSTGRES_DSN", "postgres://fleet:fleet@localhost:5432/fleet_events?sslmode=disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnv("REDIS_DB", "0"),
		},
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		},
		Worker: WorkerConfig{
			Concurrency: getEnvAsInt("WORKER_CONCURRENCY", 4),
			Prefetch:    getEnvAsInt("WORKER_PREFETCH", 4),
		},
		DeviceAPIKey: getEnv("DEVICE_API_KEY", "dev-api-key"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func getEnvAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsedValue
}
