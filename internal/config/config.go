package config

import "os"

type Config struct {
	HTTP HTTPConfig
}

type HTTPConfig struct {
	Addr string
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr: getEnv("HTTP_ADDR", ":8080"),
		},
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
