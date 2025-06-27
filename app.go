package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"autoclipsend/logger"
	"autoclipsend/version"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/sys/windows/registry"
)

// MedalTVSettings represents the structure of MedalTV's settings.json
type MedalTVSettings struct {
	Recorder struct {
		ClipFolder string `json:"clipFolder"`
	} `json:"recorder"`
}

// MedalTVClip represents a single clip entry in Medal TV's clips.json
type MedalTVClip struct {
	UUID        string  `json:"uuid"`
	ClipID      string  `json:"clipID"`
	Status      string  `json:"Status"`
	FilePath    string  `json:"FilePath"`
	Image       string  `json:"Image"`
	GameTitle   string  `json:"GameTitle"`
	TimeCreated float64 `json:"TimeCreated"`
	ClipType    string  `json:"clipType"`
	Content     struct {
		ContentTitle       string  `json:"contentTitle"`
		VideoLengthSeconds float64 `json:"videoLengthSeconds"`
		LocalContentURL    string  `json:"localContentUrl"`
		ThumbnailURL       string  `json:"thumbnailUrl"`
		State              struct {
			Type        string `json:"type"`
			IsSuccess   bool   `json:"isSuccess"`
			IsShareable bool   `json:"isShareable"`
		} `json:"state"`
	} `json:"Content"`
}

// ClipDisplayData represents clip data optimized for frontend display
type ClipDisplayData struct {
	UUID         string  `json:"uuid"`
	Title        string  `json:"title"`
	GameTitle    string  `json:"gameTitle"`
	TimeCreated  int64   `json:"timeCreated"`
	Duration     float64 `json:"duration"`
	Thumbnail    string  `json:"thumbnail"`
	ThumbnailURL string  `json:"thumbnailUrl"`
	FilePath     string  `json:"filePath"`
	Status       string  `json:"status"`
}

// MedalTVClipsData represents the structure of Medal TV's clips.json
type MedalTVClipsData struct {
	Clips []MedalTVClip `json:"clips"`
}

// NVIDIAGallerySettings represents the structure of NVIDIA's GallerySettings.json
type NVIDIAGallerySettings struct {
	Settings struct {
		CurrentDirectoryV2 string `json:"currentDirectoryV2"`
	} `json:"settings"`
}

// App struct
type App struct {
	ctx                 context.Context
	watchers            map[string]*fsnotify.Watcher // Multiple watchers for different paths
	watcherMutex        sync.Mutex                   // Protects watcher access
	config              *Config
	configManager       *ConfigManager // Kept for backward compatibility
	isVisible           bool           // Tracks if window is visible
	startTime           time.Time      // Track when app started
	isMonitoring        bool           // Track monitoring status
	monitoredPaths      []string       // List of currently monitored paths
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
	UseMedalTV   bool   `json:"useMedalTV"`
	UseNVIDIA    bool   `json:"useNVIDIA"`
	UseCustom    bool   `json:"useCustom"`
	MedalTVPath  string `json:"medalTVPath"`
	NVIDIAPath   string `json:"nvidiaPath"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Create config manager
	configManager := NewConfigManager()

	// Load config from config manager
	config, err := configManager.LoadConfig()
	if err != nil {
		logger.Warn("Failed to load config: %v, using defaults", err)
		// LoadConfig now always returns a default config, so this shouldn't happen
	}

	// Ensure config is not nil (safety check)
	if config == nil {
		logger.Error("Config is nil, creating default config")
		config = &Config{
			WebhookURL:            "",
			MonitorPath:           `E:\Highlights\Clips\Screen Recording`,
			MaxFileSize:           10,
			CheckInterval:         2,
			StartupInitialization: true,
			WindowsStartup:        false,
			RecursiveMonitoring:   false,
			DesktopShortcut:       false,
			UseMedalTVPath:        false,
			UseNVIDIAPath:         false,
			UseCustomPath:         false,
		}
	}
	app := &App{
		config:         config,
		configManager:  configManager,
		startTime:      time.Now(),
		isMonitoring:   false,
		watchers:       make(map[string]*fsnotify.Watcher),
		monitoredPaths: make([]string, 0),
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

	// Add debug logging to check configuration
	logger.Info("=== STARTUP DEBUG INFO ===")
	logger.Info("StartupInitialization: %v", a.config.StartupInitialization)
	logger.Info("UseMedalTVPath: %v", a.config.UseMedalTVPath)
	logger.Info("UseNVIDIAPath: %v", a.config.UseNVIDIAPath)
	logger.Info("UseCustomPath: %v", a.config.UseCustomPath)
	logger.Info("MonitorPath: %s", a.config.MonitorPath)

	// Test Medal TV path detection
	if medalPath, err := a.GetMedalTVClipFolder(); err == nil {
		logger.Info("Medal TV path detected: %s", medalPath)
	} else {
		logger.Info("Medal TV path error: %v", err)
	}

	logger.Info("=== END STARTUP DEBUG INFO ===")

	// Start file watcher in a goroutine only if startup initialization is enabled
	if a.config.StartupInitialization {
		go a.startFileWatcher()
	} else {
		logger.Info("StartupInitialization is disabled - file watcher not started automatically")
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
	err := a.configManager.SaveConfig(a.config)

	// Emit event to notify frontend of config changes
	if err == nil {
		runtime.EventsEmit(a.ctx, "config-updated")
	}

	return err
}

// startFileWatcher starts monitoring the specified directories
func (a *App) startFileWatcher() {
	a.watcherMutex.Lock()
	defer a.watcherMutex.Unlock()

	// Check if monitoring is already running
	if a.isMonitoring {
		return
	}

	// Get all paths to monitor
	pathsToMonitor := a.getActivePaths()
	if len(pathsToMonitor) == 0 {
		logger.Info("No paths configured for monitoring")
		return
	}

	a.isMonitoring = true
	a.monitoredPaths = pathsToMonitor

	defer func() {
		a.stopAllWatchers()
		a.isMonitoring = false
		a.monitoredPaths = make([]string, 0)
	}()

	// Create watchers for each path
	for _, path := range pathsToMonitor {
		if err := a.createWatcherForPath(path); err != nil {
			logger.Error("Failed to create watcher for path %s: %v", path, err)
			continue
		}
	}

	if len(a.watchers) == 0 {
		logger.Error("No watchers could be created")
		return
	}

	logger.Info("File monitoring started for %d paths: %v", len(a.watchers), pathsToMonitor)

	// Release the mutex before entering the monitoring loop
	a.watcherMutex.Unlock()

	// Monitor all watchers
	a.monitorWatchers()

	a.watcherMutex.Lock()
}

// getActivePaths returns all paths that should be monitored
func (a *App) getActivePaths() []string {
	var paths []string

	if a.config.UseMedalTVPath {
		if medalPath, err := a.GetMedalTVClipFolder(); err == nil && medalPath != "" {
			logger.Info("Adding MedalTV path to monitoring: %s", medalPath)
			paths = append(paths, medalPath)
		} else {
			logger.Warn("MedalTV path enabled but could not get path: %v", err)
		}
	}

	if a.config.UseNVIDIAPath {
		if nvidiaPath, err := a.GetNVIDIACurrentDirectory(); err == nil && nvidiaPath != "" {
			logger.Info("Adding NVIDIA path to monitoring: %s", nvidiaPath)
			paths = append(paths, nvidiaPath)
		} else {
			logger.Warn("NVIDIA path enabled but could not get path: %v", err)
		}
	}

	if a.config.UseCustomPath && a.config.MonitorPath != "" {
		logger.Info("Adding custom path to monitoring: %s", a.config.MonitorPath)
		paths = append(paths, a.config.MonitorPath)
	}

	logger.Info("Total active paths for monitoring: %d", len(paths))
	return paths
}

// createWatcherForPath creates a watcher for a specific path
func (a *App) createWatcherForPath(path string) error {
	logger.Info("Creating watcher for path: %s", path)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher for %s: %v", path, err)
	}

	// Add the directory to watch
	err = watcher.Add(path)
	if err != nil {
		watcher.Close()
		return fmt.Errorf("error adding path %s to watcher: %v", path, err)
	}

	logger.Info("Successfully added main path %s to watcher", path)

	// If recursive monitoring is enabled, add all subdirectories
	if a.config.RecursiveMonitoring {
		logger.Info("Recursive monitoring enabled - scanning subdirectories for %s", path)
		err = a.addSubdirectoriesToWatcher(watcher, path)
		if err != nil {
			logger.Warn("Error adding subdirectories for %s: %v", path, err)
			// Don't fail the entire operation - just warn
		}
	} else {
		logger.Info("Recursive monitoring disabled - watching only %s", path)
	}

	a.watchers[path] = watcher
	logger.Info("Created watcher for path: %s (recursive: %v)", path, a.config.RecursiveMonitoring)
	return nil
}

// addSubdirectoriesToWatcher recursively adds all subdirectories to a specific watcher
func (a *App) addSubdirectoriesToWatcher(watcher *fsnotify.Watcher, root string) error {
	if watcher == nil {
		return errors.New("watcher is not initialized")
	}

	dirCount := 0
	maxDirs := 10000 // Limit to prevent system overload

	logger.Info("Starting recursive directory scan for: %s", root)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Warn("Error accessing path %s: %v", path, err)
			return nil // Continue walking despite errors
		}

		if info.IsDir() && path != root {
			dirCount++
			if dirCount > maxDirs {
				logger.Warn("Reached maximum directory limit (%d) for recursive monitoring in %s", maxDirs, root)
				return filepath.SkipDir
			}

			if watcher == nil {
				return errors.New("watcher became nil during operation")
			}

			err = watcher.Add(path)
			if err != nil {
				logger.Error("Error adding subdirectory %s to watcher: %v", path, err)
				// Continue despite errors - don't fail the entire operation
			} else {
				logger.Debug("Added subdirectory to watch: %s", path)
			}
		}
		return nil
	})

	logger.Info("Recursive scan completed for %s: %d directories added", root, dirCount)
	return err
}

// monitorWatchers handles events from all watchers
func (a *App) monitorWatchers() {
	// Create channels to merge all watcher events
	events := make(chan fsnotify.Event, 1000) // Increased buffer size
	errors := make(chan error, 100)

	// Start goroutines for each watcher
	for path, watcher := range a.watchers {
		go func(w *fsnotify.Watcher, p string) {
			logger.Debug("Started monitoring goroutine for path: %s", p)
			for {
				select {
				case event, ok := <-w.Events:
					if !ok {
						logger.Debug("Watcher events channel closed for path: %s", p)
						return
					}
					events <- event
				case err, ok := <-w.Errors:
					if !ok {
						logger.Debug("Watcher errors channel closed for path: %s", p)
						return
					}
					errors <- fmt.Errorf("error from watcher %s: %v", p, err)
				}
			}
		}(watcher, path)
	}

	logger.Info("Started monitoring %d paths with enhanced event handling", len(a.watchers))

	// Main monitoring loop
	eventCount := 0
	for a.isMonitoring {
		select {
		case event := <-events:
			eventCount++
			if eventCount%100 == 0 {
				logger.Debug("Processed %d events so far", eventCount)
			}
			a.handleWatcherEvent(event)
		case err := <-errors:
			logger.Error("Watcher error: %v", err)
		case <-time.After(100 * time.Millisecond):
			// Small timeout to prevent busy waiting
		}
	}

	logger.Info("Monitoring stopped after processing %d events", eventCount)
}

// handleWatcherEvent processes a file system event
func (a *App) handleWatcherEvent(event fsnotify.Event) {
	logger.Info("File system event: %s - %s", event.Op, event.Name)

	if event.Op&fsnotify.Create == fsnotify.Create {
		// If it's a directory and recursive monitoring is enabled, add it to all relevant watchers
		if a.config.RecursiveMonitoring {
			if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
				logger.Debug("New directory detected: %s", event.Name)
				a.watcherMutex.Lock()
				addedToWatchers := 0
				for path, watcher := range a.watchers {
					// Check if the new directory is under this watched path
					if strings.HasPrefix(event.Name, path) && watcher != nil {
						err := watcher.Add(event.Name)
						if err != nil {
							logger.Error("Failed to add new directory %s to watcher %s: %v", event.Name, path, err)
						} else {
							addedToWatchers++
							logger.Debug("Added new directory %s to watcher %s", event.Name, path)
						}
					}
				}
				a.watcherMutex.Unlock()
				if addedToWatchers > 0 {
					logger.Info("Added new directory %s to %d watchers", event.Name, addedToWatchers)
				}
			}
		}

		if a.isVideoFile(event.Name) {
			// Skip compressed files to avoid processing loop
			// When we send a file to Discord, it might create a "_compressed" version
			// which would trigger another notification - we want to ignore these
			if strings.Contains(filepath.Base(event.Name), "_compressed") {
				logger.Info("Skipping compressed file: %s", event.Name)
				return
			}

			logger.Info("New video file detected: %s", event.Name)
			// Wait a bit for the file to be fully written
			time.Sleep(time.Duration(a.config.CheckInterval) * time.Second)
			a.handleNewVideo(event.Name)
		} else {
			logger.Info("Non-video file created: %s", event.Name)
		}
	}
}

// stopAllWatchers closes all active watchers
func (a *App) stopAllWatchers() {
	for path, watcher := range a.watchers {
		if watcher != nil {
			watcher.Close()
			logger.Debug("Closed watcher for path: %s", path)
		}
	}
	a.watchers = make(map[string]*fsnotify.Watcher)
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

	// Update Medal TV clips.json if this is a Medal TV clip and custom name is provided
	if a.isMedalTVClip(filePath) {
		titleToSet := customName
		if titleToSet == "" {
			titleToSet = "Untitled"
		}
		err = a.updateMedalTVClipTitle(filePath, titleToSet)
		if err != nil {
			logger.Warn("Failed to update Medal TV clip title: %v", err)
			// Don't return error, just log it as this is not critical
		}
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
	return version.FormatVersion()
}

// GetAppStatus returns the current application status
func (a *App) GetAppStatus() AppStatus {
	uptime := time.Since(a.startTime)

	// Get statistics from config
	stats := a.GetStatistics()

	// Get path information
	var medalTVPath, nvidiaPath string
	if a.config.UseMedalTVPath {
		medalTVPath, _ = a.GetMedalTVClipFolder()
	}
	if a.config.UseNVIDIAPath {
		nvidiaPath, _ = a.GetNVIDIACurrentDirectory()
	}
	return AppStatus{
		Uptime:       formatDuration(uptime),
		IsMonitoring: a.isMonitoring,
		MonitorPath:  a.config.MonitorPath,
		VideosSent:   stats.TotalClips,   // Use total clips from storage
		AudiosSent:   stats.SessionClips, // Use session clips for audio count
		Version:      version.FormatVersion(),
		UseMedalTV:   a.config.UseMedalTVPath,
		UseNVIDIA:    a.config.UseNVIDIAPath,
		UseCustom:    a.config.UseCustomPath,
		MedalTVPath:  medalTVPath,
		NVIDIAPath:   nvidiaPath,
	}
}

// SaveConfig saves the entire configuration
func (a *App) SaveConfig(config Config) error {
	a.config = &config
	err := a.configManager.SaveConfig(a.config)

	// Emit event to notify frontend of config changes
	if err == nil {
		runtime.EventsEmit(a.ctx, "config-updated")
	}

	return err
}

// UpdateMonitorPath updates the monitor path and restarts watcher
func (a *App) UpdateMonitorPath(path string) error {
	a.watcherMutex.Lock()
	defer a.watcherMutex.Unlock()

	// Stop current watchers if running
	a.stopAllWatchersLocked()

	// Update config
	a.config.MonitorPath = path
	err := a.configManager.SaveConfig(a.config)
	if err != nil {
		return err
	}

	// Emit event to notify frontend of config changes
	runtime.EventsEmit(a.ctx, "config-updated")

	// Restart watchers with new configuration
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

// RestartMonitoring restarts file monitoring with current configuration
func (a *App) RestartMonitoring() error {
	logger.Info("Restarting monitoring with current configuration")

	a.watcherMutex.Lock()
	defer a.watcherMutex.Unlock()

	// Stop current watchers
	a.stopAllWatchersLocked()

	// Start new monitoring
	go a.startFileWatcher()

	logger.Info("Monitoring restart initiated")
	return nil
}

// StopMonitoring stops the file monitoring
func (a *App) StopMonitoring() error {
	a.watcherMutex.Lock()
	defer a.watcherMutex.Unlock()

	a.stopAllWatchersLocked()
	return nil
}

// GetMonitoredPaths returns the currently monitored paths
func (a *App) GetMonitoredPaths() []string {
	a.watcherMutex.Lock()
	defer a.watcherMutex.Unlock()

	return append([]string(nil), a.monitoredPaths...) // Return a copy
}

// stopAllWatchersLocked stops all watchers (assumes mutex is already locked)
func (a *App) stopAllWatchersLocked() {
	a.isMonitoring = false
	for path, watcher := range a.watchers {
		if watcher != nil {
			watcher.Close()
			logger.Debug("Closed watcher for path: %s", path)
		}
	}
	a.watchers = make(map[string]*fsnotify.Watcher)
	a.monitoredPaths = make([]string, 0)
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

// GetVersionInfo returns detailed version information
func (a *App) GetVersionInfo() map[string]string {
	return version.GetDetailedVersionInfo()
}

// GetBuildInfo returns the build information
func (a *App) GetBuildInfo() version.BuildInfo {
	return version.GetBuildInfo()
}

// CheckForUpdates checks for available updates on GitHub
func (a *App) CheckForUpdates() version.UpdateInfo {
	// GitHub repository for auto-send-clips
	githubRepo := "Beelzebub2/auto-send-clips"
	return version.CheckForUpdates(githubRepo)
}

// OpenUpdateURL opens the update URL in the default browser
func (a *App) OpenUpdateURL(url string) error {
	if url == "" {
		return errors.New("no update URL provided")
	}

	// Use Windows-specific command to open URL
	cmd := exec.Command("cmd", "/c", "start", url)
	return cmd.Run()
}

// CreateDesktopShortcut creates a desktop shortcut for the application
func (a *App) CreateDesktopShortcut() error {
	// Windows process creation flags
	const CREATE_NO_WINDOW = 0x08000000

	// Get the current executable path
	exePath, err := os.Executable()
	if err != nil {
		logger.Error("failed to get executable path: %v", err)
		return errors.New("failed to get executable path")
	}

	// Get the desktop path using PowerShell with completely hidden execution
	psGetDesktopScript := `[Environment]::GetFolderPath('Desktop')`
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-Command", psGetDesktopScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NO_WINDOW,
	}
	desktopBytes, err := cmd.Output()
	if err != nil {
		logger.Error("failed to get desktop path: %v", err)
		return errors.New("failed to get desktop path")
	}

	desktopPath := strings.TrimSpace(string(desktopBytes))
	if desktopPath == "" {
		// Fallback to default path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error("failed to get user home directory: %v", err)
			return errors.New("failed to get user home directory")
		}
		desktopPath = filepath.Join(homeDir, "Desktop")
	}

	// Ensure the desktop directory exists
	if err := os.MkdirAll(desktopPath, 0755); err != nil {
		logger.Error("failed to create desktop directory: %v", err)
		return errors.New("failed to create desktop directory")
	}

	shortcutPath := filepath.Join(desktopPath, "AutoClipSend.lnk")

	// Create PowerShell script to create the shortcut using proper escaping
	psScript := fmt.Sprintf(`
$WshShell = New-Object -comObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut('%s')
$Shortcut.TargetPath = '%s'
$Shortcut.WorkingDirectory = '%s'
$Shortcut.Description = 'AutoClipSend - Automatic clip sender to Discord'
$Shortcut.Save()
`, shortcutPath, exePath, filepath.Dir(exePath))

	// Execute the PowerShell script with completely hidden execution
	cmd = exec.Command("powershell", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NO_WINDOW,
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("failed to create desktop shortcut: %v, output: %s", err, string(output))
		return errors.New("failed to create desktop shortcut: " + string(output))
	}

	logger.Info("Desktop shortcut created successfully at: %s", shortcutPath)
	return nil
}

// RemoveDesktopShortcut removes the desktop shortcut
func (a *App) RemoveDesktopShortcut() error {
	// Windows process creation flags
	const CREATE_NO_WINDOW = 0x08000000

	// Get the desktop path using PowerShell to get the actual Desktop folder location
	psGetDesktopScript := `[Environment]::GetFolderPath('Desktop')`
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-Command", psGetDesktopScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NO_WINDOW,
	}
	desktopBytes, err := cmd.Output()
	if err != nil {
		logger.Error("failed to get desktop path: %v", err)
		return errors.New("failed to get desktop path")
	}

	desktopPath := strings.TrimSpace(string(desktopBytes))
	if desktopPath == "" {
		// Fallback to default path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error("failed to get user home directory: %v", err)
			return errors.New("failed to get user home directory")
		}
		desktopPath = filepath.Join(homeDir, "Desktop")
	}

	shortcutPath := filepath.Join(desktopPath, "AutoClipSend.lnk")

	// Check if shortcut exists
	if _, err := os.Stat(shortcutPath); os.IsNotExist(err) {
		logger.Info("Desktop shortcut does not exist, nothing to remove")
		return nil // Shortcut doesn't exist, nothing to remove
	}

	// Remove the shortcut
	err = os.Remove(shortcutPath)
	if err != nil {
		logger.Error("failed to remove desktop shortcut: %v", err)
		return errors.New("failed to remove desktop shortcut" + err.Error())
	}

	logger.Info("Desktop shortcut removed successfully from: %s", shortcutPath)
	return nil
}

// HasDesktopShortcut checks if a desktop shortcut exists
func (a *App) HasDesktopShortcut() bool {
	// Windows process creation flags
	const CREATE_NO_WINDOW = 0x08000000

	// Get the desktop path using PowerShell to get the actual Desktop folder location
	psGetDesktopScript := `[Environment]::GetFolderPath('Desktop')`
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-Command", psGetDesktopScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: CREATE_NO_WINDOW,
	}
	desktopBytes, err := cmd.Output()
	if err != nil {
		// Fallback to default path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		desktopPath := filepath.Join(homeDir, "Desktop")
		shortcutPath := filepath.Join(desktopPath, "AutoClipSend.lnk")
		_, err = os.Stat(shortcutPath)
		return err == nil
	}

	desktopPath := strings.TrimSpace(string(desktopBytes))
	if desktopPath == "" {
		// Fallback to default path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		desktopPath = filepath.Join(homeDir, "Desktop")
	}

	shortcutPath := filepath.Join(desktopPath, "AutoClipSend.lnk")
	_, err = os.Stat(shortcutPath)
	return err == nil
}

// SetDesktopShortcut enables or disables desktop shortcut
func (a *App) SetDesktopShortcut(enabled bool) error {
	a.config.DesktopShortcut = enabled

	if enabled {
		err := a.CreateDesktopShortcut()
		if err != nil {
			a.config.DesktopShortcut = false // Revert on error
			logger.Error("failed to create desktop shortcut: %v", err)
			return errors.New("failed to create desktop shortcut")
		}
	} else {
		err := a.RemoveDesktopShortcut()
		if err != nil {
			logger.Error("failed to remove desktop shortcut: %v", err)
			return errors.New("failed to remove desktop shortcut")
		}
	}

	return a.configManager.SaveConfig(a.config)
}

// GetMedalTVClipFolder reads the clipFolder path from MedalTV's settings.json
func (a *App) GetMedalTVClipFolder() (string, error) {
	// Get user's AppData directory
	appDataPath := os.Getenv("APPDATA")
	if appDataPath == "" {
		return "", errors.New("APPDATA environment variable not found")
	}

	// Construct path to MedalTV settings file
	medalSettingsPath := filepath.Join(appDataPath, "Medal", "store", "settings.json")

	// Check if file exists
	if _, err := os.Stat(medalSettingsPath); os.IsNotExist(err) {
		return "", errors.New("MedalTV settings file not found - is MedalTV installed?")
	}

	// Read the file
	data, err := os.ReadFile(medalSettingsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read MedalTV settings: %v", err)
	}

	// Parse JSON
	var settings MedalTVSettings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return "", fmt.Errorf("failed to parse MedalTV settings: %v", err)
	}

	clipFolder := settings.Recorder.ClipFolder
	if clipFolder == "" {
		return "", errors.New("clipFolder not found in MedalTV settings")
	}

	// Verify the path exists
	if _, err := os.Stat(clipFolder); os.IsNotExist(err) {
		return "", fmt.Errorf("MedalTV clip folder does not exist: %s", clipFolder)
	}

	return clipFolder, nil
}

// GetNVIDIACurrentDirectory reads the currentDirectoryV2 path from NVIDIA's GallerySettings.json
func (a *App) GetNVIDIACurrentDirectory() (string, error) {
	// Get user's Local AppData directory
	localAppDataPath := os.Getenv("LOCALAPPDATA")
	if localAppDataPath == "" {
		return "", errors.New("LOCALAPPDATA environment variable not found")
	}

	// Construct path to NVIDIA settings file
	nvidiaSettingsPath := filepath.Join(localAppDataPath, "NVIDIA Corporation", "NVIDIA Overlay", "GallerySettings.json")

	// Check if file exists
	if _, err := os.Stat(nvidiaSettingsPath); os.IsNotExist(err) {
		return "", errors.New("NVIDIA GallerySettings file not found - is NVIDIA Overlay installed?")
	}

	// Read the file
	data, err := os.ReadFile(nvidiaSettingsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read NVIDIA settings: %v", err)
	}

	// Parse JSON
	var settings NVIDIAGallerySettings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return "", fmt.Errorf("failed to parse NVIDIA settings: %v", err)
	}

	currentDirectory := settings.Settings.CurrentDirectoryV2
	if currentDirectory == "" {
		return "", errors.New("currentDirectoryV2 not found in NVIDIA settings")
	}

	// Verify the path exists
	if _, err := os.Stat(currentDirectory); os.IsNotExist(err) {
		return "", fmt.Errorf("NVIDIA current directory does not exist: %s", currentDirectory)
	}

	return currentDirectory, nil
}

// isMedalTVClip checks if a file is from Medal TV by comparing its path with the Medal TV clip folder
func (a *App) isMedalTVClip(filePath string) bool {
	if !a.config.UseMedalTVPath {
		return false
	}

	medalTVPath, err := a.GetMedalTVClipFolder()
	if err != nil {
		return false
	}

	// Normalize paths for comparison
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}

	absMedalPath, err := filepath.Abs(medalTVPath)
	if err != nil {
		return false
	}

	// Check if the file is within the Medal TV clip folder
	return strings.HasPrefix(absFilePath, absMedalPath)
}

// updateMedalTVClipTitle updates the contentTitle in Medal TV's clips.json file for a specific clip
func (a *App) updateMedalTVClipTitle(filePath, customTitle string) error {
	// Get Medal TV clips.json path
	appDataPath := os.Getenv("APPDATA")
	if appDataPath == "" {
		return errors.New("APPDATA environment variable not found")
	}

	clipsJSONPath := filepath.Join(appDataPath, "Medal", "store", "clips.json")

	// Check if file exists
	if _, err := os.Stat(clipsJSONPath); os.IsNotExist(err) {
		return errors.New("Medal TV clips.json file not found")
	}

	// Read the file
	data, err := os.ReadFile(clipsJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read clips.json: %v", err)
	}

	// Parse JSON as generic map to preserve structure
	var clipsData map[string]interface{}
	err = json.Unmarshal(data, &clipsData)
	if err != nil {
		return fmt.Errorf("failed to parse clips.json: %v", err)
	}

	// Get the clips array
	clips, ok := clipsData["clips"].([]interface{})
	if !ok {
		return errors.New("clips array not found in clips.json")
	}

	// Normalize the file path for comparison
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Find the clip with matching localContentUrl and update its contentTitle
	updated := false
	for _, clip := range clips {
		clipMap, ok := clip.(map[string]interface{})
		if !ok {
			continue
		}

		content, ok := clipMap["Content"].(map[string]interface{})
		if !ok {
			continue
		}

		localContentURL, ok := content["localContentUrl"].(string)
		if !ok {
			continue
		}

		// Normalize the local content URL for comparison
		absLocalURL, err := filepath.Abs(localContentURL)
		if err != nil {
			continue
		}

		// Check if this is the clip we're looking for
		if absFilePath == absLocalURL {
			// Update the content title
			if customTitle != "" {
				content["contentTitle"] = customTitle
				content["hasTitle"] = true
			} else {
				content["contentTitle"] = "Untitled"
				content["hasTitle"] = false
			}
			updated = true
			logger.Info("Updated Medal TV clip title for %s to: %s", filepath.Base(filePath), customTitle)
			break
		}
	}

	if !updated {
		logger.Warn("Could not find clip in clips.json for file: %s", filePath)
		return nil // Don't treat this as an error, just log it
	}

	// Write the updated data back to the file
	updatedData, err := json.MarshalIndent(clipsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated clips.json: %v", err)
	}

	err = os.WriteFile(clipsJSONPath, updatedData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated clips.json: %v", err)
	}

	return nil
}

// GetMedalTVClips reads and returns all clips from Medal TV's clips.json file
func (a *App) GetMedalTVClips() ([]ClipDisplayData, error) {
	// Get Medal TV clips.json path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	appDataPath := filepath.Join(homeDir, "AppData", "Roaming")
	clipsJSONPath := filepath.Join(appDataPath, "Medal", "store", "clips.json")

	// Check if file exists
	if _, err := os.Stat(clipsJSONPath); os.IsNotExist(err) {
		return nil, errors.New("Medal TV clips.json file not found")
	}

	// Read the file
	data, err := os.ReadFile(clipsJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read clips.json: %v", err)
	}

	// Parse the JSON as a map of clips
	var clipsMap map[string]MedalTVClip
	if err := json.Unmarshal(data, &clipsMap); err != nil {
		return nil, fmt.Errorf("failed to parse clips.json: %v", err)
	}

	// Convert to display data and sort by time (latest first)
	var clips []ClipDisplayData
	for uuid, clip := range clipsMap {
		// Only include clips with proper file paths
		if clip.FilePath == "" {
			continue
		}

		// Check if video file exists
		if _, err := os.Stat(clip.FilePath); os.IsNotExist(err) {
			continue
		}

		// Determine the display title
		title := clip.Content.ContentTitle
		if title == "" {
			title = clip.GameTitle
		}
		if title == "" {
			title = "Untitled Clip"
		}

		clipData := ClipDisplayData{
			UUID:         uuid,
			Title:        title,
			GameTitle:    clip.GameTitle,
			TimeCreated:  int64(clip.TimeCreated),
			Duration:     clip.Content.VideoLengthSeconds,
			Thumbnail:    clip.Image,
			ThumbnailURL: clip.Content.ThumbnailURL,
			FilePath:     clip.FilePath,
			Status:       clip.Status,
		}
		clips = append(clips, clipData)
	}

	// Sort clips by time created (latest first)
	for i := 0; i < len(clips)-1; i++ {
		for j := i + 1; j < len(clips); j++ {
			if clips[i].TimeCreated < clips[j].TimeCreated {
				clips[i], clips[j] = clips[j], clips[i]
			}
		}
	}

	return clips, nil
}

// SendClipToDiscord sends a specific clip to Discord
func (a *App) SendClipToDiscord(clipUUID string) error {
	// Get the clip data
	clips, err := a.GetMedalTVClips()
	if err != nil {
		return fmt.Errorf("failed to get clips: %v", err)
	}

	// Find the specific clip
	var targetClip *ClipDisplayData
	for _, clip := range clips {
		if clip.UUID == clipUUID {
			targetClip = &clip
			break
		}
	}

	if targetClip == nil {
		return errors.New("clip not found")
	}

	// Check if file exists
	if _, err := os.Stat(targetClip.FilePath); os.IsNotExist(err) {
		return errors.New("clip file not found")
	}

	// Use existing SendToDiscord function with custom name
	customName := fmt.Sprintf("%s_%d", targetClip.Title, targetClip.TimeCreated)
	return a.SendToDiscord(targetClip.FilePath, customName, false)
}
