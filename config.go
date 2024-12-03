package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	ConfigFilePath string
	FragmentsPath  string `yaml:"fragments"`
	PagePath       string `yaml:"pages"`
	BuildPath      string `yaml:"build"`
}

func GetConfiguration(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal([]byte(data), cfg)
	if err != nil {
		return nil, err
	}

	cfg.ConfigFilePath = path
	return cfg, nil
}
