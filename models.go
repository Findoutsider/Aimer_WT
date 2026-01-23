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

type Mod struct {
	Title        string          `json:"title"`
	Author       string          `json:"author"`
	Note         string          `json:"note"`
	LinkBiliBili string          `json:"link_bilibili"`
	LinkWtLive   string          `json:"link_wtlive"`
	LinkVideo    string          `json:"link_video"`
	Language     []string        `json:"language"`
	SizeStr      string          `json:"size_str"`
	Capabilities map[string]bool `json:"capabilities"`
}

// Manifest 结构体定义
type Manifest struct {
	InstalledMods map[string]ModInfo `json:"installed_mods"`
	FileMap       map[string]string  `json:"file_map"`
}

type ModInfo struct {
	Files       []string `json:"files"`
	InstallTime string   `json:"install_time"`
}

type FolderType string

const (
	GameFolder      FolderType = "game"
	PendingFolder   FolderType = "pending"
	VoiceFolder     FolderType = "voice"
	GameVoiceFolder FolderType = "game_voice"
	SkinFolder      FolderType = "skin"
	GunScopeFolder  FolderType = "gunscope"
)

type FolderPath string

var (
	root                = "data"
	PendingFolderPath   = FolderPath(filepath.Join(root, "pending"))
	VoiceFolderPath     = FolderPath(filepath.Join(root, "voice"))
	GameVoiceFolderPath = FolderPath(filepath.Join(gamePath, "sound/mod"))
	SkinFolderPath      = FolderPath(filepath.Join(gamePath, "UserSkins"))
	GunScopeFolderPath  = FolderPath(filepath.Join(root, "gunscope"))
)

var FolderPaths = map[FolderType]FolderPath{
	GameFolder:      FolderPath(gamePath),
	PendingFolder:   PendingFolderPath,
	VoiceFolder:     VoiceFolderPath,
	GameVoiceFolder: GameVoiceFolderPath,
	SkinFolder:      SkinFolderPath,
	GunScopeFolder:  GunScopeFolderPath,
}

var folders = []FolderPath{
	PendingFolderPath,
	VoiceFolderPath,
	GameVoiceFolderPath,
	SkinFolderPath,
	GunScopeFolderPath,
}

func GetPath(fType FolderType) string {
	switch fType {
	case GameFolder:
		return gamePath
	case SkinFolder:
		if gamePath == "" {
			return filepath.Join(root, "skin")
		}
		return filepath.Join(gamePath, "UserSkins")
	case VoiceFolder:
		return filepath.Join(root, "voice")
	case GameVoiceFolder:
		return filepath.Join(gamePath, "sound/mod")
	case PendingFolder:
		return filepath.Join(root, "pending")
	case GunScopeFolder:
		return filepath.Join(root, "gunscope")
	default:
		return root
	}
}
