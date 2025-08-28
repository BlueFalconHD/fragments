package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SiteRoot      string
	FragmentsPath string `yaml:"fragments"`
	PagePath      string `yaml:"pages"`
	IncludePath   string `yaml:"include"`
	BuildPath     string `yaml:"build"`
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

	// Get the root of the site, the directory in which the config file (provided path) is located
	cfg.SiteRoot = filepath.Dir(path)
	return cfg, nil
}
