package main

import (
	"embed"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS
var vp *viper.Viper
var basePath string
var gamePath string

func main() {
	InitLogger(basePath)
	go startLogRotationTrigger(rotator)
	go InitAppFolders()

	exePath, _ := os.Executable()
	basePath = filepath.Dir(exePath)
	InitConfig()
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Aimer_WT",
		Width:     1200,
		Height:    740,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func startLogRotationTrigger(l *lumberjack.Logger) {
	for {
		// 计算距离明天零点还有多久
		now := time.Now()
		next := now.Add(time.Hour * 24)
		next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
		t := time.NewTimer(next.Sub(now))

		<-t.C
		l.Rotate() // 强制手动切割文件，生成带日期的新文件
	}
}
