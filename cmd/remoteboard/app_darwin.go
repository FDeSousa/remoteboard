//go:build darwin

package main

import (
	"log"

	"github.com/getlantern/systray"
)

// runApp starts the macOS menu-bar icon and blocks until the user quits.
// Must be called from the main goroutine on macOS.
func runApp(tabletURL, cfgPath string) {
	systray.Run(func() {
		systray.SetTitle("⌨ RemoteBoard")
		systray.SetTooltip("RemoteBoard is running")

		mURL := systray.AddMenuItem("Tablet URL: "+tabletURL, "Copy the tablet connection URL")
		mConfig := systray.AddMenuItem("Open Config File", "Open buttons.json in the default editor")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit RemoteBoard", "Stop the server and exit")

		go func() {
			for {
				select {
				case <-mURL.ClickedCh:
					copyToClipboard(tabletURL)
				case <-mConfig.ClickedCh:
					openFile(cfgPath)
				case <-mQuit.ClickedCh:
					systray.Quit()
				}
			}
		}()
	}, func() {
		log.Println("remoteboard: exiting")
	})
}
