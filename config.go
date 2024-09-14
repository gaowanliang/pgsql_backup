package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

// loadConfig reads the YAML configuration file
func loadConfig(configPath string) (*Config, error) {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
