package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Port     int      `yaml:"port"`
	Backends []string `yaml:"backends"`
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
