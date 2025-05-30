package main

import (
	"fmt"
	"time"

	"github.com/getlantern/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// InitTray initializes the system tray icon and menu
func (a *App) InitTray() {
	// IMPORTANT: Call systray.Run in a goroutine so it doesn't block
	go func() {
		fmt.Println("Starting system tray...")
		systray.Run(a.onTrayReady, a.onTrayExit)
		fmt.Println("System tray stopped - this should only happen on exit")
	}()
}

// onTrayReady sets up the tray icon and menu
func (a *App) onTrayReady() {
	fmt.Println("Tray icon is now ready")

	// Use the icon from main.go
	systray.SetIcon(icon)
	systray.SetTitle("AutoClipSend")
	systray.SetTooltip("AutoClipSend (Right-click for options)")

	// Create menu items
	mStatus := systray.AddMenuItem("AutoClipSend is monitoring...", "Status")
	mStatus.Disable() // This item is just for display

	systray.AddSeparator()

	mShow := systray.AddMenuItem("Show Window", "Show the main window")

	systray.AddSeparator()

	mExit := systray.AddMenuItem("Exit", "Exit the application completely")

	// Handle menu item clicks
	go func() {
		fmt.Println("Starting tray menu event handler")
		for {
			select {
			case <-mShow.ClickedCh:
				fmt.Println("Show window clicked in tray menu")
				// Use ShowFromTray to properly restore the window
				a.ShowFromTray()

			case <-mExit.ClickedCh:
				fmt.Println("Exit clicked in tray menu - shutting down app completely")
				// First quit the systray
				systray.Quit()
				// Then quit the entire application
				if a.ctx != nil {
					runtime.Quit(a.ctx)
				}
				return
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
		fmt.Println("MinimizeToTray - Hiding window and setting isVisible=false")
		// Set window as not visible first
		a.isVisible = false

		// Make sure to hide the window with both contexts for reliability
		runtime.WindowHide(a.ctx)

		// Tell the frontend that we've minimized to tray
		runtime.EventsEmit(a.ctx, "app-minimized-to-tray")

		fmt.Println("App is now minimized to tray - tray icon should be visible")
	} else {
		fmt.Println("ERROR: MinimizeToTray called with nil context!")
	}
}

// ShowFromTray shows the app from the system tray
func (a *App) ShowFromTray() {
	if a.ctx == nil {
		fmt.Println("ERROR: ShowFromTray called with nil context!")
		return
	}

	fmt.Println("ShowFromTray - Showing window from tray and bringing to front")

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

	fmt.Println("Window is now visible and brought to front")
}

// ToggleVisibility toggles the app's visibility
func (a *App) ToggleVisibility() {
	if a.isVisible {
		a.MinimizeToTray()
	} else {
		a.ShowFromTray()
	}
}
