package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// ConfigManager handles saving and loading configuration
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".autoclipsend")
	os.MkdirAll(configDir, 0755)

	return &ConfigManager{
		configPath: filepath.Join(configDir, "config.json"),
	}
}

// SaveConfig saves the configuration to file
func (cm *ConfigManager) SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.configPath, data, 0644)
}

// LoadConfig loads the configuration from file
func (cm *ConfigManager) LoadConfig() (*Config, error) {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		// Return default config if file doesn't exist
		return &Config{
			MonitorPath:   `E:\Highlights\Clips\Screen Recording`,
			MaxFileSize:   10 * 1024 * 1024, // 10MB
			CheckInterval: 2,
		}, nil
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
