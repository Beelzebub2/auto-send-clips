package main

import (
	"time"

	"autoclipsend/logger"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// InitTray initializes the system tray icon and menu
func (a *App) InitTray() {
	// IMPORTANT: Call systray.Run in a goroutine so it doesn't block
	go func() {
		logger.Info("Starting system tray...")
		systray.Run(a.onTrayReady, a.onTrayExit)
		logger.Info("System tray stopped - this should only happen on exit")
	}()
}

// onTrayReady sets up the tray icon and menu
func (a *App) onTrayReady() {
	logger.Debug("Tray icon is now ready")

	// Use the icon from main.go
	systray.SetIcon(icon)
	systray.SetTitle("AutoClipSend")
	systray.SetTooltip("AutoClipSend (Right-click for menu)")

	// Status Section
	mStatus := systray.AddMenuItem("⚡ AutoClipSend", "Status")
	mStatus.Disable() // Make it non-clickable

	mStatusState := systray.AddMenuItem("● Monitoring Active", "Current status")
	mStatusState.Disable() // Make it non-clickable

	// First Section Separator
	systray.AddSeparator()

	// Main Actions Section
	mShow := systray.AddMenuItem("👁️ Show Window", "Open the main window")
	mToggleMonitoring := systray.AddMenuItemCheckbox("▶️ Pause Monitoring", "Pause/Resume file monitoring", true)

	// Final Section Separator
	systray.AddSeparator()

	// Exit
	mExit := systray.AddMenuItem("❌ Exit", "Exit the application completely")

	// Status update goroutine (refreshes status periodically)
	go func() {
		for {
			if a.isMonitoring {
				mStatusState.SetTitle("● Monitoring Active")
				mToggleMonitoring.Check()
				mToggleMonitoring.SetTitle("⏸️ Pause Monitoring")
			} else {
				mStatusState.SetTitle("○ Monitoring Paused")
				mToggleMonitoring.Uncheck()
				mToggleMonitoring.SetTitle("▶️ Resume Monitoring")
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Set click handlers for menu items
	go func() {
		for {
			select {
			case <-mShow.ClickedCh:
				logger.Debug("Show window clicked in tray menu")
				a.ShowFromTray()

			case <-mToggleMonitoring.ClickedCh:
				a.isMonitoring = !a.isMonitoring
				runtime.EventsEmit(a.ctx, "toggle-monitoring", a.isMonitoring)

			case <-mExit.ClickedCh:
				logger.Info("Exit clicked in tray menu - shutting down app completely")
				systray.Quit()
				if a.ctx != nil {
					runtime.Quit(a.ctx)
				}
			}
		}
	}()
}

// onTrayExit is called when the tray is exiting
func (a *App) onTrayExit() {
	// Optional cleanup logic
}

// MinimizeToTray minimizes the app to the system tray
func (a *App) MinimizeToTray() {
	if a.ctx != nil {
		logger.Debug("MinimizeToTray - Hiding window and setting isVisible=false")
		// Set window as not visible first
		a.isVisible = false

		// Make sure to hide the window with both contexts for reliability
		runtime.WindowHide(a.ctx)

		// Tell the frontend that we've minimized to tray
		runtime.EventsEmit(a.ctx, "app-minimized-to-tray")

		logger.Debug("App is now minimized to tray - tray icon should be visible")
	} else {
		logger.Error("MinimizeToTray called with nil context!")
	}
}

// ShowFromTray shows the app from the system tray
func (a *App) ShowFromTray() {
	if a.ctx == nil {
		logger.Error("ShowFromTray called with nil context!")
		return
	}

	logger.Debug("ShowFromTray - Showing window from tray and bringing to front")

	// Set always on top before showing/unminimizing
	runtime.WindowSetAlwaysOnTop(a.ctx, true)

	// Show and unminimize the window
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)

	// Show again to ensure visibility (sometimes needed on Windows)
	runtime.WindowShow(a.ctx)

	time.Sleep(200 * time.Millisecond)
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	// Update visibility state
	a.isVisible = true

	// Emit event to notify frontend
	runtime.EventsEmit(a.ctx, "app-restored-from-tray")

	logger.Debug("Window is now visible and brought to front")
}

// ToggleVisibility toggles the app's visibility
func (a *App) ToggleVisibility() {
	if a.isVisible {
		a.MinimizeToTray()
	} else {
		a.ShowFromTray()
	}
}

// ShowCustomTrayMenu shows a custom dark-themed context menu
func (a *App) ShowCustomTrayMenu() {
	if a.ctx == nil {
		return
	}

	// Show a small custom menu window at a reasonable position
	runtime.WindowShow(a.ctx)
	runtime.WindowSetSize(a.ctx, 200, 150)
	runtime.WindowSetAlwaysOnTop(a.ctx, true)

	// Emit event to show custom tray menu in frontend
	runtime.EventsEmit(a.ctx, "show-tray-menu")
}
