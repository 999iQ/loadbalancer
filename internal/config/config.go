package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type Config struct {
	Port                     int           `yaml:"port"`
	Backends                 []string      `yaml:"backends"`
	ServerShutdownTimeoutSec time.Duration `yaml:"server_shutdown_timeout_sec"`
}

type RateLimitConfig struct {
	Default struct {
		Capacity int           `yaml:"capacity"`
		FillRate time.Duration `yaml:"fill_rate"` // скорость пополнения токенов
	} `yaml:"default"`
	// настройки RateLimit для  отдельных IP клиентов
	Overrides map[string]struct {
		Capacity int           `yaml:"capacity"`
		FillRate time.Duration `yaml:"fill_rate"`
	} `yaml:"overrides"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}

func LoadRateLimitConfig(path string) (*RateLimitConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg RateLimitConfig
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}
