package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Stats represents application statistics
type Stats struct {
	TotalClips     int       `json:"total_clips"`
	LastClipTime   time.Time `json:"last_clip_time"`
	SessionClips   int       `json:"session_clips"`
	TotalSize      int64     `json:"total_size_bytes"`
	StartTime      time.Time `json:"start_time"`
	LastUpdateTime time.Time `json:"last_update_time"`
}

// Config holds application configuration and statistics
type Config struct {
	// Settings
	WebhookURL            string `json:"webhook_url"`
	DiscordWebhook        string `json:"discord_webhook"` // Alternative name for webhook
	MonitorPath           string `json:"monitor_path"`
	MaxFileSize           int64  `json:"max_file_size"`          // in MB
	CheckInterval         int    `json:"check_interval"`         // in seconds
	StartupInitialization bool   `json:"startup_initialization"` // Whether to start monitoring on startup
	WindowsStartup        bool   `json:"windows_startup"`        // Whether to start with Windows
	RecursiveMonitoring   bool   `json:"recursive_monitoring"`   // Whether to monitor subfolders recursively

	// Statistics
	Stats
}

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

// LoadConfig loads the configuration from file (legacy)
func (cm *ConfigManager) LoadConfig() (*Config, error) {
	data, err := os.ReadFile(cm.configPath)
	if err != nil { // Return default config if file doesn't exist
		return &Config{
			MonitorPath:           `E:\Highlights\Clips\Screen Recording`,
			MaxFileSize:           10, // 10MB
			CheckInterval:         2,
			StartupInitialization: true,  // Default to enabled
			WindowsStartup:        false, // Default to disabled
			RecursiveMonitoring:   false, // Default to disabled
			Stats: Stats{
				TotalClips:     0,
				SessionClips:   0,
				TotalSize:      0,
				StartTime:      time.Now(),
				LastUpdateTime: time.Now(),
			},
		}, nil
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// IncrementClipCount increments the clip counters and updates file size
func (cm *ConfigManager) IncrementClipCount(config *Config, fileSize int64) error {
	config.TotalClips++
	config.SessionClips++
	config.TotalSize += fileSize
	config.LastClipTime = time.Now()
	config.LastUpdateTime = time.Now()
	return cm.SaveConfig(config)
}

// ResetSessionStats resets session-specific statistics
func (cm *ConfigManager) ResetSessionStats(config *Config) error {
	config.SessionClips = 0
	config.StartTime = time.Now()
	config.LastUpdateTime = time.Now()
	return cm.SaveConfig(config)
}

// GetUptime returns the uptime since start time
func (cm *ConfigManager) GetUptime(config *Config) time.Duration {
	return time.Since(config.StartTime)
}
