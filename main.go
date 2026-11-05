package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"keuangan_nawawi/database"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// setupLogFile membuat log file di folder AppData aplikasi
func setupLogFile() *os.File {
	if runtime.GOOS != "windows" {
		return nil
	}
	appDir := database.GetAppDataDir()
	logPath := filepath.Join(appDir, "keuangan_nawawi.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil
	}
	log.SetOutput(f)
	return f
}

func main() {
	// Setup log file untuk Windows
	if logFile := setupLogFile(); logFile != nil {
		defer logFile.Close()
	}

	// Tangkap panic agar error bisa dilihat user
	defer func() {
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("Aplikasi mengalami error fatal:\n%v", r)
			log.Println(errMsg)

			// Tulis error ke folder AppData
			if runtime.GOOS == "windows" {
				appDir := database.GetAppDataDir()
				errPath := filepath.Join(appDir, "ERROR.txt")
				os.WriteFile(errPath, []byte(errMsg), 0644)
			}
		}
	}()

	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "Sistem Akuntansi Keuangan Nawawi",
		Width:            1280,
		Height:           720,
		MinWidth:         1024,
		MinHeight:        600,
		WindowStartState: options.Maximised,
		Frameless:        false,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId:               "nawawi-keuangan-app-6c2e88a3-294b",
			OnSecondInstanceLaunch: app.onSecondInstanceLaunch,
		},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown: func(ctx context.Context) {
			_ = database.CloseDB()
			os.Exit(0)
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		log.Println("Error:", err.Error())
	}
	os.Exit(0)
}

