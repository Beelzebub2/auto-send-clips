package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds application configuration
type Config struct {
	WebhookURL        string `json:"webhook_url"`
	DiscordWebhook    string `json:"discord_webhook"` // Alternative name for webhook
	MonitorPath       string `json:"monitor_path"`
	MaxFileSize       int64  `json:"max_file_size"`  // in MB
	CheckInterval     int    `json:"check_interval"` // in seconds
	AudioExtraction   bool   `json:"audio_extraction"`
	ShowNotifications bool   `json:"show_notifications"`
	AutoCompress      bool   `json:"auto_compress"`
}

// ConfigManager handles saving and loading configuration (legacy)
// Note: This is kept for backward compatibility, new code should use storage.go
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

// LoadConfig loads the configuration from file (legacy)
func (cm *ConfigManager) LoadConfig() (*Config, error) {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		// Return default config if file doesn't exist
		return &Config{
			MonitorPath:       `E:\Highlights\Clips\Screen Recording`,
			MaxFileSize:       10, // 10MB
			CheckInterval:     2,
			AudioExtraction:   false,
			ShowNotifications: true,
			AutoCompress:      true,
		}, nil
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
