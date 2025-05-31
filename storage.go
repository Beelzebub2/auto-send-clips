package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AppData represents the persistent application data
type AppData struct {
	Settings   Config `json:"settings"`
	Statistics Stats  `json:"statistics"`
	mu         sync.RWMutex
}

// Stats represents application statistics
type Stats struct {
	TotalClips     int       `json:"total_clips"`
	LastClipTime   time.Time `json:"last_clip_time"`
	SessionClips   int       `json:"session_clips"`
	TotalSize      int64     `json:"total_size_bytes"`
	StartTime      time.Time `json:"start_time"`
	LastUpdateTime time.Time `json:"last_update_time"`
}

var (
	appData      *AppData
	dataPath     string
	settingsFile string
	once         sync.Once
)

// InitStorage initializes the storage system
func InitStorage() error {
	var initErr error
	once.Do(func() {
		// Get %APPDATA% directory
		appDataDir := os.Getenv("APPDATA")
		if appDataDir == "" {
			initErr = fmt.Errorf("APPDATA environment variable not found")
			return
		}

		// Create AutoClipSend directory
		dataPath = filepath.Join(appDataDir, "AutoClipSend")
		err := os.MkdirAll(dataPath, 0755)
		if err != nil {
			initErr = fmt.Errorf("failed to create data directory: %v", err)
			return
		}

		settingsFile = filepath.Join(dataPath, "settings.json")

		// Initialize appData
		appData = &AppData{
			Settings: Config{
				MonitorPath:    `E:\Highlights\Clips\Screen Recording`,
				DiscordWebhook: "",
				MaxFileSize:    10, // Default 20 MB
			},
			Statistics: Stats{
				TotalClips:     0,
				SessionClips:   0,
				TotalSize:      0,
				StartTime:      time.Now(),
				LastUpdateTime: time.Now(),
			},
		}

		// Load existing data
		initErr = loadAppData()
		if initErr != nil {
			// If file doesn't exist, create it with defaults
			if os.IsNotExist(initErr) {
				initErr = saveAppData()
			}
		}

		// Update session start time
		appData.mu.Lock()
		appData.Statistics.StartTime = time.Now()
		appData.Statistics.SessionClips = 0
		appData.mu.Unlock()
	})

	return initErr
}

// loadAppData loads data from the settings file
func loadAppData() error {
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return err
	}

	appData.mu.Lock()
	defer appData.mu.Unlock()

	return json.Unmarshal(data, appData)
}

// saveAppData saves data to the settings file
func saveAppData() error {
	appData.mu.RLock()
	data, err := json.MarshalIndent(appData, "", "  ")
	appData.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(settingsFile, data, 0644)
}

// GetSettings returns the current settings
func GetSettings() Config {
	if appData == nil {
		InitStorage()
	}

	appData.mu.RLock()
	defer appData.mu.RUnlock()
	return appData.Settings
}

// SaveSettings saves the settings
func SaveSettings(config Config) error {
	if appData == nil {
		InitStorage()
	}

	appData.mu.Lock()
	appData.Settings = config
	appData.Statistics.LastUpdateTime = time.Now()
	appData.mu.Unlock()

	return saveAppData()
}

// GetStatistics returns the current statistics
func GetStatistics() Stats {
	if appData == nil {
		InitStorage()
	}

	appData.mu.RLock()
	defer appData.mu.RUnlock()
	return appData.Statistics
}

// IncrementClipCount increments the clip counters
func IncrementClipCount(fileSize int64) error {
	if appData == nil {
		InitStorage()
	}

	appData.mu.Lock()
	appData.Statistics.TotalClips++
	appData.Statistics.SessionClips++
	appData.Statistics.TotalSize += fileSize
	appData.Statistics.LastClipTime = time.Now()
	appData.Statistics.LastUpdateTime = time.Now()
	appData.mu.Unlock()

	return saveAppData()
}

// ResetSessionStats resets session-specific statistics
func ResetSessionStats() error {
	if appData == nil {
		InitStorage()
	}

	appData.mu.Lock()
	appData.Statistics.SessionClips = 0
	appData.Statistics.StartTime = time.Now()
	appData.Statistics.LastUpdateTime = time.Now()
	appData.mu.Unlock()

	return saveAppData()
}

// GetDataPath returns the application data directory path
func GetDataPath() string {
	if dataPath == "" {
		InitStorage()
	}
	return dataPath
}

// ExportSettings exports settings to a JSON file
func ExportSettings(filePath string) error {
	if appData == nil {
		InitStorage()
	}

	appData.mu.RLock()
	data, err := json.MarshalIndent(appData, "", "  ")
	appData.mu.RUnlock()

	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// ImportSettings imports settings from a JSON file
func ImportSettings(filePath string) error {
	if appData == nil {
		InitStorage()
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var importedData AppData
	err = json.Unmarshal(data, &importedData)
	if err != nil {
		return err
	}

	appData.mu.Lock()
	// Only import settings, keep current statistics
	appData.Settings = importedData.Settings
	appData.Statistics.LastUpdateTime = time.Now()
	appData.mu.Unlock()

	return saveAppData()
}

// GetStorageInfo returns information about the storage system
func GetStorageInfo() map[string]interface{} {
	if appData == nil {
		InitStorage()
	}

	info := make(map[string]interface{})
	info["data_path"] = dataPath
	info["settings_file"] = settingsFile

	// Check if files exist
	if _, err := os.Stat(settingsFile); err == nil {
		info["settings_file_exists"] = true
		if stat, err := os.Stat(settingsFile); err == nil {
			info["settings_file_size"] = stat.Size()
			info["settings_file_modified"] = stat.ModTime()
		}
	} else {
		info["settings_file_exists"] = false
	}

	appData.mu.RLock()
	info["total_clips"] = appData.Statistics.TotalClips
	info["session_clips"] = appData.Statistics.SessionClips
	info["total_size_mb"] = float64(appData.Statistics.TotalSize) / (1024 * 1024)
	appData.mu.RUnlock()

	return info
}
