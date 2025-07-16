package config

import "os"

type AppConfig struct {
	AppPort  string
	LogLevel string
}

type PostgresConfig struct {
	DBHost      string
	DBPort      string
	DBName      string
	DBUser      string
	DBPassword  string
	PostgresDSN string
}

type Config struct {
	App      AppConfig
	Postgres PostgresConfig
}

func LoadConfig() *Config {
	cfg := &Config{
		App: AppConfig{
			AppPort:  getEnv("APP_PORT", "8080"),
			LogLevel: getEnv("LOG_LEVEL", "DEBUG"),
		},
		Postgres: PostgresConfig{
			DBHost:      getEnv("DB_HOST", "db"),
			DBPort:      getEnv("DB_PORT", "5432"),
			DBName:      getEnv("DB_NAME", "subtracker"),
			DBUser:      getEnv("DB_USER", "postgres"),
			DBPassword:  getEnv("DB_PASSWORD", "supersecret"),
			PostgresDSN: getEnv("POSTGRES_DSN", "postgres://postgres:supersecret@db:5432/subtracker?sslmode=disable"),
		},
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
