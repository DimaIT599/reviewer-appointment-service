package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	DataBase `yaml:"postgres"`
	Server   `yaml:"server"`
}

type DataBase struct {
	Host      string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	User      string `yaml:"user" env:"DB_USER" env-default:"postgres"`
	Password  string `yaml:"password" env:"DB_PASSWORD" env-default:"postgres"`
	Data_base string `yaml:"database" env:"DB_NAME" env-default:"reviewer_appointment"`
	PortDB    int    `yaml:"port" env:"DB_PORT" env-default:"5432"`
}

type Server struct {
	PortServer int `yaml:"port" env:"SERVER_PORT" env-default:"8081"`
}

func MustConfig(config_path string) *Config {
	var cfg Config

	// Если CONFIG_PATH пустой или не указан, используем только переменные окружения
	useConfigFile := config_path != "" && config_path != `""`

	if useConfigFile {
		// Сначала читаем YAML (если файл существует)
		if _, err := os.Stat(config_path); err == nil {
			if err := cleanenv.ReadConfig(config_path, &cfg); err != nil {
				log.Printf("Warning: failed to read config file: %v", err)
			}
		}
	}

	// Читаем переменные окружения (они всегда имеют приоритет)
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Printf("Warning: failed to read environment variables: %v", err)
	}

	// Переменные окружения имеют ВЫСШИЙ приоритет - перезаписываем после чтения YAML
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		cfg.Host = dbHost
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		cfg.User = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		cfg.Password = dbPassword
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Data_base = dbName
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		var port int
		if _, err := fmt.Sscanf(dbPort, "%d", &port); err == nil {
			cfg.PortDB = port
		} else {
			log.Printf("Warning: invalid DB_PORT value: %s", dbPort)
		}
	}
	if serverPort := os.Getenv("SERVER_PORT"); serverPort != "" {
		var port int
		if _, err := fmt.Sscanf(serverPort, "%d", &port); err == nil {
			cfg.PortServer = port
		} else {
			log.Printf("Warning: invalid SERVER_PORT value: %s", serverPort)
		}
	}

	log.Printf("Config loaded - DB: host=%s, port=%d, user=%s, database=%s (env: DB_HOST=%s, DB_PORT=%s, useConfigFile=%v)",
		cfg.Host, cfg.PortDB, cfg.User, cfg.Data_base,
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), useConfigFile)

	return &cfg
}
