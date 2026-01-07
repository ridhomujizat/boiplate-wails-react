package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"onx-screen-record/internal/app"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

const deepLinkScheme = "onxrecord://"

// sendToWebhook sends deep link URL to webhook for testing/debugging
func sendToWebhook(deepLinkURL string) {
	webhookURL := "https://webhook.site/ff296acd-5d4e-4f9a-b7e8-fb36a8e65316"
	// Parse the URL to extract data/token
	parsed, err := url.Parse(deepLinkURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		fmt.Println("=========================")
		return
	}

	fmt.Printf("Scheme: %s\n", parsed.Scheme)
	fmt.Printf("Host: %s\n", parsed.Host)
	fmt.Printf("Path: %s\n", parsed.Path)

	// Extract query parameters
	query := parsed.Query()
	email := query.Get("email")
	token := query.Get("token")
	payload := map[string]string{
		"url":   deepLinkURL,
		"email": email,
		"token": token,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal webhook payload: %v\n", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to send webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Webhook sent successfully. Status: %d\n", resp.StatusCode)
}

func main() {
	app := app.NewApp()

	// Handle deep link from initial launch
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, deepLinkScheme) {
			sendToWebhook(arg)
			app.SetInitialDeepLink(arg)
			break
		}
	}

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
				// Check for deep link URL in command line arguments
				for _, arg := range secondInstanceData.Args {
					if strings.HasPrefix(arg, deepLinkScheme) {
						sendToWebhook(arg)
						app.HandleDeepLink(arg)
						break
					}
				}
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
		Mac: &mac.Options{
			OnUrlOpen: func(url string) {
				sendToWebhook(url)
				app.HandleDeepLink(url)
			},
		},
	}

	err := wails.Run(wailsOption)

	if err != nil {
		println("Error:", err.Error())
	}
}
