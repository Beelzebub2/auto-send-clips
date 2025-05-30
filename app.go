package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx           context.Context
	watcher       *fsnotify.Watcher
	config        *Config
	configManager *ConfigManager
	isVisible     bool // Tracks if window is visible
}

// NewApp creates a new App application struct
func NewApp() *App {
	configManager := NewConfigManager()
	config, err := configManager.LoadConfig()
	if err != nil {
		config = &Config{
			MonitorPath:   `E:\Highlights\Clips\Screen Recording`,
			MaxFileSize:   10 * 1024 * 1024, // 10MB
			CheckInterval: 2,
		}
	}

	return &App{
		config:        config,
		configManager: configManager,
	}
}

// startup is called when the app starts. The context here
// can be used to call the frontend via the application binding.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.isVisible = true

	// Initialize the system tray first to ensure it's available
	a.InitTray()

	// Start file watcher in a goroutine
	go a.startFileWatcher()
}

// domReady is called when the DOM is ready, just before the frontend shows
func (a *App) domReady(ctx context.Context) {
	// Called when DOM is ready
}

// beforeClose is called when the window is trying to close
func (a *App) beforeClose(ctx context.Context) bool {
	fmt.Println("==========================================")
	fmt.Println("beforeClose called - window is trying to close")

	// Just minimize to tray instead of closing
	a.MinimizeToTray()

	fmt.Println("Window minimized to tray, returning false to prevent app from closing")
	fmt.Println("==========================================")

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
		fmt.Printf("Error creating watcher: %v\n", err)
		return
	}
	defer a.watcher.Close()

	// Add the directory to watch
	err = a.watcher.Add(a.config.MonitorPath)
	if err != nil {
		fmt.Printf("Error adding path to watcher: %v\n", err)
		return
	}

	fmt.Printf("Watching directory: %s\n", a.config.MonitorPath)

	for {
		select {
		case event, ok := <-a.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if a.isVideoFile(event.Name) {
					fmt.Printf("New video file detected: %s\n", event.Name)
					// Wait a bit for the file to be fully written
					time.Sleep(time.Duration(a.config.CheckInterval) * time.Second)
					a.handleNewVideo(event.Name)
				}
			}
		case err, ok := <-a.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

// ShowNotification triggers a notification for a new video
func (a *App) ShowNotification(fileName, filePath string) {
	fmt.Printf("Showing notification for: %s\n", fileName)

	// Always show the window using the same logic as the tray menu
	a.ShowFromTray()
	// Give the window a moment to appear before sending the notification
	time.Sleep(250 * time.Millisecond)

	// Now emit the event to the frontend to show the notification
	runtime.EventsEmit(a.ctx, "newVideoDetected", map[string]string{
		"fileName": fileName,
		"filePath": filePath,
	})

	// Ensure window is brought to foreground with a toggle of always-on-top
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
	time.Sleep(100 * time.Millisecond)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)

	// Minimize to tray again after notification
	// a.MinimizeToTray() // <-- Commented out to allow user interaction
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
	fmt.Println("HandleWindowClose called from frontend")
	// Just delegate to MinimizeToTray to ensure consistent behavior
	fmt.Println("Minimizing to tray...")
	a.MinimizeToTray()
	fmt.Println("Finished minimizing to tray")
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
		return fmt.Errorf("webhook URL not set")
	}

	// Check file size
	_, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("error getting file info: %v", err)
	}

	var finalPath string
	var cleanup bool

	if audioOnly {
		// Extract audio from video
		finalPath, err = a.extractAudio(filePath)
		if err != nil {
			return fmt.Errorf("error extracting audio: %v", err)
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
		return fmt.Errorf("error getting final file info: %v", err)
	}

	if finalInfo.Size() > a.config.MaxFileSize {
		// Compress the file
		compressedPath, err := a.compressFile(finalPath, audioOnly)
		if err != nil {
			return fmt.Errorf("error compressing file: %v", err)
		}
		finalPath = compressedPath
		cleanup = true
		defer func() {
			if cleanup {
				os.Remove(finalPath)
			}
		}()
	}

	// Send to Discord
	return a.sendFileToDiscord(finalPath, customName)
}

// sendFileToDiscord sends the file to Discord via webhook
func (a *App) sendFileToDiscord(filePath, customName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add the file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("error creating form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("error copying file: %v", err)
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
		return fmt.Errorf("error closing writer: %v", err)
	}

	// Send the request
	req, err := http.NewRequest("POST", a.config.WebhookURL, &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
