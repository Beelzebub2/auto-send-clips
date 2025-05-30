package main

import (
	"fmt"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// NotificationHandler manages the sending of notifications from Go to the frontend
type NotificationHandler struct {
	app *App
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(app *App) *NotificationHandler {
	return &NotificationHandler{
		app: app,
	}
}

// SendVideoNotification sends a notification for a new video to the frontend
func (nh *NotificationHandler) SendVideoNotification(fileName, filePath string) {
	// Exit early if context is nil
	if nh.app.ctx == nil {
		fmt.Println("ERROR: Cannot send notification - context is nil")
		return
	}

	fmt.Printf("SendVideoNotification: Sending notification for %s\n", fileName)

	// Define the payload
	payload := map[string]string{
		"fileName": fileName,
		"filePath": filePath,
	}

	// First, ensure the window is visible
	nh.app.ShowFromTray()
	time.Sleep(500 * time.Millisecond) // Increased wait time

	// Log before emitting event
	fmt.Println("About to emit newVideoDetected event with payload:", payload)

	// Emit the event multiple times with delays to ensure it's caught
	for i := 0; i < 3; i++ {
		fmt.Printf("Attempt %d: Emitting newVideoDetected event\n", i+1)
		runtime.EventsEmit(nh.app.ctx, "newVideoDetected", payload)
		time.Sleep(200 * time.Millisecond)
	}

	// Always bring window to front and make it visible
	runtime.WindowShow(nh.app.ctx)
	runtime.WindowSetAlwaysOnTop(nh.app.ctx, true)
	time.Sleep(200 * time.Millisecond)
	runtime.WindowSetAlwaysOnTop(nh.app.ctx, false)
}

// TestNotification sends a test notification
func (nh *NotificationHandler) TestNotification() {
	if nh.app.ctx == nil {
		fmt.Println("ERROR: Cannot send test notification - context is nil")
		return
	}

	fmt.Println("Sending test notification")
	nh.SendVideoNotification("TestVideo.mp4", "C:\\Test\\TestVideo.mp4")
}
