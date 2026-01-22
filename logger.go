package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var logger *log.Logger
var appCtx context.Context // 全局保存 ctx 供日志使用
var rotator *lumberjack.Logger

// InitLogger 初始化本地文件日志
func InitLogger(basePath string) {
	logDir := filepath.Join(basePath, "logs")
	_ = os.MkdirAll(logDir, 0755)

	// 初始化滚动配置
	rotator = &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    5,  // 5MB
		MaxBackups: 10, // 保留10个旧文件
		MaxAge:     30, // 保留30天
		Compress:   false,
		LocalTime:  true,
	}

	// 设置输出：控制台 + 滚动文件
	multi := io.MultiWriter(os.Stdout, rotator)

	// 初始化全局 Logger
	logger = log.New(multi, "", log.LstdFlags|log.Lshortfile)
}

// SetLogContext 在 Wails OnStartup 时调用，传入 ctx
func SetLogContext(ctx context.Context) {
	appCtx = ctx
}

// Scan 专门用于推送扫描进度，只更新前端最后一行，不写本地文件
func Scan(format string, v ...any) {
	message := fmt.Sprintf(format, v...)

	// 只推送到前端，不占用磁盘 IO
	if appCtx != nil {
		runtime.EventsEmit(appCtx, "ev_update_scan", message)
	}
}

func Log(level string, message string) {
	fullMsg := fmt.Sprintf("[%s] %s", level, message)

	if logger != nil {
		logger.Println(fullMsg)
	}

	if appCtx != nil {
		runtime.EventsEmit(appCtx, "ev_append_log", fullMsg)
	}
}

func Info(format string, v ...any) {
	Log("INFO", fmt.Sprintf(format, v...))
}

func Warn(format string, v ...any) {
	Log("WARN", fmt.Sprintf(format, v...))
}

func Success(format string, v ...any) {
	Log("SUCCESS", fmt.Sprintf(format, v...))
}

func Error(format string, v ...any) {
	Log("ERROR", fmt.Sprintf(format, v...))
}
