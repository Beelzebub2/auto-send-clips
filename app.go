package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"autoclipsend/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows/registry"
)

// App struct
type App struct {
	ctx                 context.Context
	watcher             *fsnotify.Watcher
	config              *Config
	configManager       *ConfigManager // Kept for backward compatibility
	isVisible           bool           // Tracks if window is visible
	startTime           time.Time      // Track when app started
	isMonitoring        bool           // Track monitoring status
	notificationHandler *NotificationHandler
	// Note: videosSent and audiosSent moved to persistent storage
}

// AppStatus represents the current application status
type AppStatus struct {
	Uptime       string `json:"uptime"`
	IsMonitoring bool   `json:"isMonitoring"`
	MonitorPath  string `json:"monitorPath"`
	VideosSent   int    `json:"videosSent"`
	AudiosSent   int    `json:"audiosSent"`
	Version      string `json:"version"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Create config manager
	configManager := NewConfigManager()

	// Load config from config manager
	config, err := configManager.LoadConfig()
	if err != nil {
		logger.Warn("Failed to load config: %v, using defaults", err)
		// Create default config if loading fails
		config = &Config{
			MonitorPath:           `E:\Highlights\Clips\Screen Recording`,
			MaxFileSize:           10, // 10MB
			CheckInterval:         2,
			StartupInitialization: true,
			WindowsStartup:        false, // Default to disabled
			Stats: Stats{
				TotalClips:     0,
				SessionClips:   0,
				TotalSize:      0,
				StartTime:      time.Now(),
				LastUpdateTime: time.Now(),
			},
		}
	}

	app := &App{
		config:        config,
		configManager: configManager,
		startTime:     time.Now(),
		isMonitoring:  false,
	}

	// Create notification handler after app is initialized
	app.notificationHandler = NewNotificationHandler(app)
	logger.Info("Application initialized with config: monitor_path=%s, max_file_size=%dMB",
		config.MonitorPath, config.MaxFileSize)

	return app
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.isVisible = true
	// Initialize the system tray first to ensure it's available
	a.InitTray()

	// Start file watcher in a goroutine only if startup initialization is enabled
	if a.config.StartupInitialization {
		go a.startFileWatcher()
	}
}

// domReady is called when the DOM is ready
func (a *App) domReady(ctx context.Context) {
	logger.Debug("DOM ready event received")
	// Called when DOM is ready
}

// beforeClose is called when the window is trying to close
func (a *App) beforeClose(ctx context.Context) bool {
	logger.Info("beforeClose called - window is trying to close")

	// Just minimize to tray instead of closing
	a.MinimizeToTray()

	logger.Info("Window minimized to tray, returning false to prevent app from closing")

	// Return false to prevent the application from closing
	return false
}

// GetConfig returns the current configuration
func (a *App) GetConfig() *Config {
	return a.config
}

// SetWebhookURL sets the Discord webhook URL
func (a *App) SetWebhookURL(url string) error {
	a.config.WebhookURL = url
	return a.configManager.SaveConfig(a.config)
}

// startFileWatcher starts monitoring the specified directory
func (a *App) startFileWatcher() {
	var err error
	a.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.Error("Error creating watcher: %v", err)
		a.isMonitoring = false
		return
	}
	defer a.watcher.Close()

	// Add the directory to watch
	err = a.watcher.Add(a.config.MonitorPath)
	if err != nil {
		logger.Error("Error adding path to watcher: %v", err)
		a.isMonitoring = false
		return
	}
	// If recursive monitoring is enabled, add all subdirectories
	if a.config.RecursiveMonitoring {
		err = a.addSubdirectories(a.config.MonitorPath)
		if err != nil {
			logger.Warn("Error adding subdirectories: %v", err)
		}
	}
	a.isMonitoring = true
	logger.Info("File monitoring started: path=%s recursive=%v", a.config.MonitorPath, a.config.RecursiveMonitoring)

	for {
		select {
		case event, ok := <-a.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				// If it's a directory and recursive monitoring is enabled, add it to watcher
				if a.config.RecursiveMonitoring {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						a.watcher.Add(event.Name)
					}
				}

				if a.isVideoFile(event.Name) {
					logger.Info("New video file detected: %s", event.Name)
					// Wait a bit for the file to be fully written
					time.Sleep(time.Duration(a.config.CheckInterval) * time.Second)
					a.handleNewVideo(event.Name)
				}
			}
		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			logger.Error("Watcher error: %v", err)
		}
	}
}

// ShowNotification triggers a notification for a new video
func (a *App) ShowNotification(fileName, filePath string) {
	logger.Info("ShowNotification called for: %s", fileName)

	// Use the dedicated notification handler
	a.notificationHandler.SendVideoNotification(fileName, filePath)
}

// GetFileSize returns the size of a file in MB
func (a *App) GetFileSize(filePath string) (float64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return float64(fileInfo.Size()) / (1024 * 1024), nil
}

// HandleWindowClose is called from JavaScript to properly handle window close
func (a *App) HandleWindowClose() {
	logger.Debug("HandleWindowClose called from frontend")
	// Just delegate to MinimizeToTray to ensure consistent behavior
	logger.Debug("Minimizing to tray...")
	a.MinimizeToTray()
	logger.Debug("Finished minimizing to tray")
}

// BringToFront brings the window to the front
func (a *App) BringToFront() {
	a.isVisible = true
	runtime.WindowShow(a.ctx)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
}

// IsVisible returns whether the window is currently visible
func (a *App) IsVisible() bool {
	return a.isVisible
}

// Minimize minimizes the window to the taskbar (not to tray)
func (a *App) Minimize() {
	runtime.WindowMinimise(a.ctx)
	// We're still visible, just minimized to taskbar
	// So we don't change isVisible flag here
}

// Maximize maximizes the window
func (a *App) Maximize() {
	runtime.WindowMaximise(a.ctx)
}

// SendToDiscord sends the file to Discord via webhook
// Moved from notification.go to app.go for correct method binding
func (a *App) SendToDiscord(filePath, customName string, audioOnly bool) error {
	if a.config.WebhookURL == "" {
		logger.Error("webhook URL not set")
		return errors.New("webhook URL not set")
	}

	// Check file size
	_, err := os.Stat(filePath)
	if err != nil {
		logger.Error("error getting file info: %v", err)
		return errors.New("error getting file info")
	}

	var finalPath string
	var cleanup bool

	if audioOnly {
		// Extract audio from video
		finalPath, err = a.extractAudio(filePath)
		if err != nil {
			logger.Error("error extracting audio: %v", err)
			return errors.New("error extracting audio")
		}
		cleanup = true
		defer func() {
			if cleanup {
				os.Remove(finalPath)
			}
		}()
	} else {
		finalPath = filePath
	}

	// Check final file size
	finalInfo, err := os.Stat(finalPath)
	if err != nil {
		logger.Error("error getting final file info: %v", err)
		return errors.New("error getting final file info")
	}

	if finalInfo.Size() > a.config.MaxFileSize*1024*1024 {
		// Compress the file
		compressedPath, err := a.compressFile(finalPath, audioOnly)
		if err != nil {
			logger.Error("error compressing file: %v", err)
			return errors.New("error compressing file")
		}
		finalPath = compressedPath
		cleanup = true
		defer func() {
			if cleanup {
				os.Remove(finalPath)
			}
		}()
	} // Send to Discord
	err = a.sendFileToDiscord(finalPath, customName)
	if err != nil {
		return err
	}

	// Get file size for statistics
	fileInfo, _ := os.Stat(finalPath)
	fileSize := int64(0)
	if fileInfo != nil {
		fileSize = fileInfo.Size()
	} // Increment clip count using ConfigManager
	err = a.configManager.IncrementClipCount(a.config, fileSize)
	if err != nil {
		logger.Warn("Failed to update clip statistics: %v", err)
	}

	return nil
}

// sendFileToDiscord sends the file to Discord via webhook
func (a *App) sendFileToDiscord(filePath, customName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		logger.Error("error opening file: %v", err)
		return errors.New("error opening file")
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add the file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		logger.Error("error creating form file: %v", err)
		return errors.New("error creating form file")
	}

	_, err = io.Copy(part, file)
	if err != nil {
		logger.Error("error copying file: %v", err)
		return errors.New("error copying file")
	}

	// Add custom message if provided
	if customName != "" {
		payload := map[string]interface{}{
			"content": customName,
		}
		payloadBytes, _ := json.Marshal(payload)
		writer.WriteField("payload_json", string(payloadBytes))
	}

	err = writer.Close()
	if err != nil {
		logger.Error("error closing writer: %v", err)
		return errors.New("error closing writer")
	}

	// Send the request
	req, err := http.NewRequest("POST", a.config.WebhookURL, &buf)
	if err != nil {
		logger.Error("error creating request: %v", err)
		return errors.New("error creating request")
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("error sending request: %v", err)
		return errors.New("error sending request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("discord API error: %d - %s", resp.StatusCode, string(body))
		return errors.New("discord API error")
	}

	return nil
}

// GetUptime returns the application's uptime
func (a *App) GetUptime() string {
	return time.Since(a.startTime).String()
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return version
}

// GetAppStatus returns the current application status
func (a *App) GetAppStatus() AppStatus {
	uptime := time.Since(a.startTime)

	// Get statistics from config
	stats := a.GetStatistics()

	return AppStatus{
		Uptime:       formatDuration(uptime),
		IsMonitoring: a.isMonitoring,
		MonitorPath:  a.config.MonitorPath,
		VideosSent:   stats.TotalClips,   // Use total clips from storage
		AudiosSent:   stats.SessionClips, // Use session clips for audio count
		Version:      version,
	}
}

// SaveConfig saves the entire configuration
func (a *App) SaveConfig(config Config) error {
	a.config = &config
	return a.configManager.SaveConfig(a.config)
}

// UpdateMonitorPath updates the monitor path and restarts watcher
func (a *App) UpdateMonitorPath(path string) error {
	// Stop current watcher if running
	if a.watcher != nil {
		a.watcher.Close()
		a.isMonitoring = false
	}

	// Update config
	a.config.MonitorPath = path
	err := a.configManager.SaveConfig(a.config)
	if err != nil {
		return err
	}

	// Restart watcher with new path
	go a.startFileWatcher()
	return nil
}

// StartMonitoring starts the file monitoring
func (a *App) StartMonitoring() error {
	if !a.isMonitoring {
		go a.startFileWatcher()
	}
	return nil
}

// StopMonitoring stops the file monitoring
func (a *App) StopMonitoring() error {
	if a.watcher != nil {
		a.watcher.Close()
		a.isMonitoring = false
	}
	return nil
}

// SelectFolder opens a folder selection dialog
func (a *App) SelectFolder() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "Select Monitor Folder",
	}

	result, err := runtime.OpenDirectoryDialog(a.ctx, options)
	if err != nil {
		return "", err
	}

	return result, nil
}

// formatDuration formats a duration into a readable string
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return intToStr(days) + "d " + intToStr(hours) + "h " + intToStr(minutes) + "m " + intToStr(seconds) + "s"
	} else if hours > 0 {
		return intToStr(hours) + "h " + intToStr(minutes) + "m " + intToStr(seconds) + "s"
	} else if minutes > 0 {
		return intToStr(minutes) + "m " + intToStr(seconds) + "s"
	}
	return intToStr(seconds) + "s"
}

func intToStr(i int) string {
	return strconv.Itoa(i)
}

// GetStatistics returns the current application statistics
func (a *App) GetStatistics() Stats {
	return Stats{
		TotalClips:     a.config.TotalClips,
		LastClipTime:   a.config.LastClipTime,
		SessionClips:   a.config.SessionClips,
		TotalSize:      a.config.TotalSize,
		StartTime:      a.config.StartTime,
		LastUpdateTime: a.config.LastUpdateTime,
	}
}

// GetStorageInfo returns information about the storage system
func (a *App) GetStorageInfo() map[string]interface{} {
	info := make(map[string]interface{})
	info["data_path"] = a.configManager.configPath
	info["settings_file"] = a.configManager.configPath

	// Check if files exist
	if _, err := os.Stat(a.configManager.configPath); err == nil {
		info["settings_file_exists"] = true
		if stat, err := os.Stat(a.configManager.configPath); err == nil {
			info["settings_file_size"] = stat.Size()
			info["settings_file_modified"] = stat.ModTime()
		}
	} else {
		info["settings_file_exists"] = false
	}

	info["total_clips"] = a.config.TotalClips
	info["session_clips"] = a.config.SessionClips
	info["total_size_mb"] = float64(a.config.TotalSize) / (1024 * 1024)

	return info
}

// ExportData exports settings and statistics to a file
func (a *App) ExportData(filePath string) error {
	data, err := json.MarshalIndent(a.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// ImportData imports settings from a file
func (a *App) ImportData(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var importedConfig Config
	err = json.Unmarshal(data, &importedConfig)
	if err != nil {
		return err
	}

	// Keep current session stats, only import settings and total stats
	importedConfig.SessionClips = a.config.SessionClips
	importedConfig.StartTime = a.config.StartTime
	importedConfig.LastUpdateTime = time.Now()

	a.config = &importedConfig
	return a.configManager.SaveConfig(a.config)
}

// ResetSessionStats resets session-specific statistics
func (a *App) ResetSessionStats() error {
	return a.configManager.ResetSessionStats(a.config)
}

// GetDataPath returns the application data directory path
func (a *App) GetDataPath() string {
	return filepath.Dir(a.configManager.configPath)
}

// SetWindowsStartup enables or disables Windows startup
func (a *App) SetWindowsStartup(enabled bool) error {
	a.config.WindowsStartup = enabled

	if enabled {
		err := a.addToWindowsStartup()
		if err != nil {
			a.config.WindowsStartup = false // Revert on error
			logger.Error("failed to add to Windows startup: %v", err)
			return errors.New("failed to add to Windows startup")
		}
	} else {
		err := a.removeFromWindowsStartup()
		if err != nil {
			logger.Error("failed to remove from Windows startup: %v", err)
			return errors.New("failed to remove from Windows startup")
		}
	}

	return a.configManager.SaveConfig(a.config)
}

func (a *App) addToWindowsStartup() error {
	exePath, err := os.Executable()
	if err != nil {
		logger.Error("failed to get executable path: %v", err)
		return errors.New("failed to get executable path")
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		logger.Error("failed to open registry key: %v", err)
		return errors.New("failed to open registry key")
	}
	defer key.Close()

	err = key.SetStringValue("AutoClipSend", exePath)
	if err != nil {
		logger.Error("failed to set registry value: %v", err)
		return errors.New("failed to set registry value")
	}

	return nil
}

func (a *App) removeFromWindowsStartup() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		logger.Error("failed to open registry key: %v", err)
		return errors.New("failed to open registry key")
	}
	defer key.Close()

	err = key.DeleteValue("AutoClipSend")
	if err != nil && err != registry.ErrNotExist {
		logger.Error("failed to delete registry value: %v", err)
		return errors.New("failed to delete registry value")
	}

	return nil
}

// IsInWindowsStartup checks if the application is currently set to start with Windows
func (a *App) IsInWindowsStartup() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue("AutoClipSend")
	return err == nil
}

// addSubdirectories recursively adds all subdirectories to the watcher
func (a *App) addSubdirectories(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != root {
			err = a.watcher.Add(path)
			if err != nil {
				logger.Error("Error adding subdirectory %s to watcher: %v", path, err)
			} else {
				logger.Debug("Added subdirectory to watch: %s", path)
			}
		}
		return nil
	})
}
