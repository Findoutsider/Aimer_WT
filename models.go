package main

import "path/filepath"

type ThemeMeta struct {
	Filename string `json:"filename"`
	Name     string `json:"name"`
	Author   string `json:"author"`
	Version  string `json:"version"`
}

type ThemeFile struct {
	Meta struct {
		Name    string `json:"name"`
		Author  string `json:"author"`
		Version string `json:"version"`
	} `json:"meta"`
	Colors map[string]any `json:"colors"`
}

type UnzipTask struct {
	Paths      []string
	TargetDir  string
	OnProgress func(current, total int, filename string) // 用于更新扫描条
	OnLog      func(level, message string)               // 用于记录持久日志
	OnFinished func()                                    // 任务完成回调
}

type ActiveTheme struct {
	ActiveTheme string `json:"active_theme"`
}

type FolderType string

const (
	GameFolder     FolderType = "game"
	PendingFolder  FolderType = "pending"
	VoiceFolder    FolderType = "voice"
	SkinFolder     FolderType = "skin"
	GunScopeFolder FolderType = "gunscope"
)

type FolderPath string

var (
	root               = "data"
	PendingFolderPath  = FolderPath(filepath.Join(root, "pending"))
	VoiceFolderPath    = FolderPath(filepath.Join(root, "voice"))
	SkinFolderPath     = FolderPath(filepath.Join(root, "skin"))
	GunScopeFolderPath = FolderPath(filepath.Join(root, "gunscope"))
)

var FolderPaths = map[FolderType]FolderPath{
	GameFolder:     FolderPath(gamePath),
	PendingFolder:  PendingFolderPath,
	VoiceFolder:    VoiceFolderPath,
	SkinFolder:     SkinFolderPath,
	GunScopeFolder: GunScopeFolderPath,
}

var folders = []FolderPath{
	PendingFolderPath,
	VoiceFolderPath,
	SkinFolderPath,
	GunScopeFolderPath,
}
