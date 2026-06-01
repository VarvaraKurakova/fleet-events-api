package config

import "os"

type Config struct {
	HTTP         HTTPConfig
	Postgres     PostgresConfig
	Redis        RedisConfig
	RabbitMQ     RabbitMQConfig
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
