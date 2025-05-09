package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"time"
)

type Config struct {
	Port                     int           `yaml:"port"`
	ServerShutdownTimeoutSec time.Duration `yaml:"server_shutdown_timeout_sec"`
	LBMethod                 string        `yaml:"lb_method"`
	Backends                 []string      `yaml:"backends"`

	RateLimit struct {
		Enabled         bool          `yaml:"enabled"`
		CleanupInterval time.Duration `yaml:"cleanup_interval"`
		Default         struct {
			RequestsPerSec int `yaml:"requests_per_sec"`
			Burst          int `yaml:"burst"`
		} `yaml:"default"`
		SpecialLimits []struct {
			IPs   []string `yaml:"ips"`
			Limit struct {
				RequestsPerSec int `yaml:"requests_per_sec"`
				Burst          int `yaml:"burst"`
			} `yaml:"limit"`
		} `yaml:"special_limits"`
	} `yaml:"rate_limit"`
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
