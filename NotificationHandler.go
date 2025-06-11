package main

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"autoclipsend/logger"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// NotificationHandler manages the sending of notifications from Go to the frontend
type NotificationHandler struct {
	app *App
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(app *App) *NotificationHandler {
	logger.Debug("Creating new notification handler")
	return &NotificationHandler{
		app: app,
	}
}

// SendVideoNotification sends a notification for a new video to the frontend
func (nh *NotificationHandler) SendVideoNotification(fileName, filePath string) {
	// Exit early if context is nil
	if nh.app.ctx == nil {
		logger.Error("Cannot send notification - context is nil")
		return
	}

	logger.Info("Sending video notification for file: %s", fileName)

	// Define the payload
	payload := map[string]string{
		"fileName": fileName,
		"filePath": filePath,
	}

	// First, ensure the window is visible
	nh.app.ShowFromTray()
	time.Sleep(500 * time.Millisecond) // Increased wait time

	logger.Debug("Emitting newVideoDetected event with payload: %+v", payload)
	// Emit the event multiple times with delays to ensure it's caught
	wailsRuntime.EventsEmit(nh.app.ctx, "newVideoDetected", payload)
	time.Sleep(200 * time.Millisecond)

	// Always bring window to front and make it visible
	wailsRuntime.WindowShow(nh.app.ctx)
	wailsRuntime.WindowSetAlwaysOnTop(nh.app.ctx, true)
	time.Sleep(200 * time.Millisecond)
	wailsRuntime.WindowSetAlwaysOnTop(nh.app.ctx, false)
}

// TestNotification sends a test notification
func (nh *NotificationHandler) TestNotification() {
	if nh.app.ctx == nil {
		logger.Error("Cannot send test notification - context is nil")
		return
	}

	logger.Info("Sending test notification")
	nh.SendVideoNotification("TestVideo.mp4", "C:\\Test\\TestVideo.mp4")
}

// Notify sends a desktop notification using the appropriate method based on the OS
func (nh *NotificationHandler) Notify(title, message string) {
	if nh.app.ctx == nil {
		logger.Error("Cannot send notification - context is nil")
		return
	}
	// Check the OS and use the appropriate command
	switch runtime.GOOS {
	case "windows":
		// Windows command
		exec.Command("powershell", "-Command", fmt.Sprintf("New-BurntToastNotification -Text '%s', '%s'", title, message)).Run()
	case "darwin":
		// macOS command
		exec.Command("osascript", "-e", fmt.Sprintf("display notification \"%s\" with title \"%s\"", message, title)).Run()
	default:
		// Fallback for other OSes (Linux, etc.), or you can add specific commands for them
		logger.Warn("Notify not implemented for this OS: %s", runtime.GOOS)
	}
}

// SendSystemNotification sends a system notification
func (nh *NotificationHandler) SendSystemNotification(title, message string) error {
	logger.Info("Sending system notification: %s - %s", title, message)

	switch runtime.GOOS {
	case "windows":
		return nh.sendWindowsNotification(title, message)
	case "darwin":
		return nh.sendMacNotification(title, message)
	case "linux":
		return nh.sendLinuxNotification(title, message)
	default:
		return errors.New("unsupported operating system: " + runtime.GOOS)
	}
}

// sendWindowsNotification sends a notification on Windows using PowerShell
func (nh *NotificationHandler) sendWindowsNotification(title, message string) error {
	script := fmt.Sprintf(`
		Add-Type -AssemblyName System.Windows.Forms
		$global:balloon = New-Object System.Windows.Forms.NotifyIcon
		$path = (Get-Process -id $pid).Path
		$balloon.Icon = [System.Drawing.Icon]::ExtractAssociatedIcon($path)
		$balloon.BalloonTipIcon = [System.Windows.Forms.ToolTipIcon]::Info
		$balloon.BalloonTipText = '%s'
		$balloon.BalloonTipTitle = '%s'
		$balloon.Visible = $true
		$balloon.ShowBalloonTip(5000)
	`, message, title)

	cmd := exec.Command("powershell", "-Command", script)
	return cmd.Run()
}

// sendMacNotification sends a notification on macOS using osascript
func (nh *NotificationHandler) sendMacNotification(title, message string) error {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// sendLinuxNotification sends a notification on Linux using notify-send
func (nh *NotificationHandler) sendLinuxNotification(title, message string) error {
	cmd := exec.Command("notify-send", title, message)
	return cmd.Run()
}

// SendVideoSystemNotification sends a system notification for a new video
func (nh *NotificationHandler) SendVideoSystemNotification(fileName, filePath string) {
	title := "AutoClipSend - New Video Detected"
	message := fmt.Sprintf("New video detected: %s\nClick to view in app.", fileName)
	err := nh.SendSystemNotification(title, message)
	if err != nil {
		logger.Error("Failed to send system notification: %v", err)
		// Fallback to in-app notification
		nh.SendVideoNotification(fileName, filePath)
	} else {
		logger.Info("System notification sent for: %s", fileName)
		// Still emit the event for the app to handle internally if needed
		if nh.app.ctx != nil {
			payload := map[string]string{
				"fileName": fileName,
				"filePath": filePath,
			}
			wailsRuntime.EventsEmit(nh.app.ctx, "newVideoDetected", payload)
		}
	}
}
