package main

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
)

type page struct {
	File    string   `yaml:"file"`
	Scripts []string `yaml:"scripts"`
}

type config struct {
	Pages      map[string]page   `yaml:"pages"`
	GlobalMeta map[string]string `yaml:"globalMeta"`
}

// readConfigFromFile reads a YAML config file and unmarshals it into a config struct.
func readConfigFromFile(file string) config {
	var cfg config

	// Read the file content
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Unmarshal YAML data into the config struct
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return cfg
}
