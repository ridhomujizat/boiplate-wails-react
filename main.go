package main

import (
	"context"
	"embed"
	"onx-screen-record/internal/app"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := app.NewApp()

	wailsOption := &options.App{
		Title:  "onx-screen-record",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: app.AppName,
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				app.ShowWindow()
			},
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			switch runtime.GOOS {
			case "darwin":
				// On macOS, hide the app when the window is closed
				app.HideWindow()
				return true // Prevent default close behavior
			case "windows":
				// On Windows, minimize to tray when the window is closed
				app.MinimizeToTray()
				return true // Prevent default close behavior
			}
			// On other OSes, minimize to tray

			return true // Prevent default close behavior
		},
	}

	err := wails.Run(wailsOption)

	if err != nil {
		println("Error:", err.Error())
	}
}
