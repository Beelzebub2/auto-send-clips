package main

import (
	"embed"
	"os/exec"

	win "golang.org/x/sys/windows"

	applogger "autoclipsend/logger"
	"autoclipsend/version"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed icon.ico
var icon []byte

// main is the entry point of the application
func main() {
	// Initialize logger
	if err := applogger.Init(); err != nil {
		// If we can't initialize the logger, we still want to show this error
		displayAlreadyRunningNotification()
		return
	}
	defer applogger.Close()

	// Log application startup
	applogger.Info("Starting AutoClipSend version %s", version.FormatVersion())

	// Prevent multiple instances using a named mutex
	mutexName, _ := win.UTF16PtrFromString("Global\\AutoClipSendMutex")
	handle, err := win.CreateMutex(nil, false, mutexName)
	if err != nil {
		applogger.Error("Error creating mutex: %v", err)
		return
	}
	lastErr := win.GetLastError()
	if lastErr == win.ERROR_ALREADY_EXISTS {
		applogger.Warn("Application instance already running")
		displayAlreadyRunningNotification()
		return
	}
	defer win.CloseHandle(handle)

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "AutoClipSend",
		Width:  430,
		Height: 550,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:  &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:         app.startup,
		OnBeforeClose:     app.beforeClose,
		HideWindowOnClose: true,                                        // Set to true to ensure window hides instead of closing
		Bind:              []interface{}{app, app.notificationHandler}, // <-- Bind the app struct and notification handler for Wails
		Frameless:         false,                                       // Use system title bar instead of custom topbar
		WindowStartState:  options.Normal,
		DisableResize:     true, // Disables window resizing
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			BackdropType:         windows.Mica,
			Theme:                windows.SystemDefault, // Use system theme
		},
		LogLevel: logger.ERROR, // Only show errors, suppress warnings and below
	})

	if err != nil {
		applogger.Error("Application error: %v", err)
	}
	applogger.Info("Application shutdown complete")
}

// displayAlreadyRunningNotification shows a Windows notification if app is already running
func displayAlreadyRunningNotification() {
	exec.Command("powershell", "-Command", "[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime]; $template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02); $textNodes = $template.GetElementsByTagName('text'); $textNodes.Item(0).AppendChild($template.CreateTextNode('AutoClipSend')); $textNodes.Item(1).AppendChild($template.CreateTextNode('AutoClipSend is already running.')); $toast = [Windows.UI.Notifications.ToastNotification]::new($template); $notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('AutoClipSend'); $notifier.Show($toast)").Start()
}
