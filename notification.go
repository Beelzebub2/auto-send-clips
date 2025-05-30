package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// NotificationApp handles the notification window
type NotificationApp struct {
	ctx      context.Context
	mainApp  *App
	fileName string
	filePath string
}

// NewNotificationApp creates a new notification app
func NewNotificationApp(mainApp *App) *NotificationApp {
	return &NotificationApp{
		mainApp: mainApp,
	}
}

// startup is called when the notification app starts
func (n *NotificationApp) startup(ctx context.Context) {
	n.ctx = ctx
}

// SetFileInfo sets the file information for the notification
func (n *NotificationApp) SetFileInfo(fileName, filePath string) {
	n.fileName = fileName
	n.filePath = filePath
}

// GetFileInfo returns the current file information
func (n *NotificationApp) GetFileInfo() map[string]string {
	return map[string]string{
		"fileName": n.fileName,
		"filePath": n.filePath,
	}
}

// SendToDiscord sends the file to Discord via the main app
func (n *NotificationApp) SendToDiscord(customName string, audioOnly bool) error {
	return n.mainApp.SendToDiscord(n.filePath, customName, audioOnly)
}

// CloseNotification closes the notification window
func (n *NotificationApp) CloseNotification() {
	runtime.Quit(n.ctx)
}

// These notification functions are now handled directly in the App struct in app.go
