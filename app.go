package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
	SetLogContext(ctx)
}

func (a *App) CloseApp() {
	runtime.Quit(a.ctx)
}

func (a *App) MinimizeApp() {
	runtime.WindowMinimise(a.ctx)
}

func (a *App) ToggleMaximise() {
	runtime.WindowToggleMaximise(a.ctx)
}

func (a *App) TogglePin(isTop bool) {
	runtime.WindowSetAlwaysOnTop(a.ctx, isTop)
}

func (a *App) SetTheme(theme string) {
	vp.Set("theme_mode", theme)
	vp.WriteConfig()
}

func (a *App) GetThemeList() []ThemeMeta {
	themesDir := filepath.Join(basePath, "themes")
	var themeList []ThemeMeta

	files, err := os.ReadDir(themesDir)
	if err != nil {
		Error("读取目录失败:", err)
		return themeList
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(themesDir, file.Name())

			t, err := ReadJSON[ThemeFile](filePath)
			if err != nil {
				logger.Printf("解析主题 %s 失败: %v", file.Name(), err)
				continue
			}

			name := t.Meta.Name
			if name == "" {
				name = file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]
			}

			author := t.Meta.Author
			if author == "" {
				author = "Unknown"
			}

			version := t.Meta.Version
			if version == "" {
				version = "1.0"
			}

			themeList = append(themeList, ThemeMeta{
				Filename: file.Name(),
				Name:     name,
				Author:   author,
				Version:  version,
			})
		}
	}

	return themeList
}

func (a *App) InitAppState() map[string]any {
	var path string
	if vp.GetString("game_path") != "" {
		path = vp.GetString("game_path")
		Info("从配置中读取游戏路径: %s", path)
	} else {
		path = GetDefaultWarThunderPath()
		Info("使用默认路径: %s", path)
	}
	theme := vp.GetString("theme_mode")

	isVerify := false
	isVerify, _ = verifyGamePath(path)
	if !isVerify {
		Warn("默认路径下不存在游戏")
	}

	return map[string]any{
		"game_path":    path,
		"path_valid":   isVerify,
		"theme":        theme,
		"active_theme": GetActiveTheme(),
		"current_mod":  vp.GetString("current_mod"),
	}
}

// SaveThemeSelection 保存主题选择
func (a *App) SaveThemeSelection(filename string) {
	vp.Set("active_theme", filename)
	vp.WriteConfig()
}

// LoadThemeContent 加载主题内容
func (a *App) LoadThemeContent(filename string) map[string]any {
	if filename == "default.json" {
		return map[string]any{"colors": nil}
	}
	themePath := filepath.Join(basePath, "themes", filename)
	t, err := ReadJSON[ThemeFile](themePath)
	if err != nil {
		Error("加载主题失败: %v", err)
		return nil
	}
	return map[string]any{"colors": t.Colors}
}

// BrowseFolder 浏览文件夹（返回路径和有效性）
func (a *App) BrowseFolder() map[string]any {
	selectedDir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择游戏安装目录",
	})
	if err != nil || selectedDir == "" {
		runtime.EventsEmit(a.ctx, "search_fail")

		Warn("未选择路径")
		return map[string]any{"path": "", "valid": false}
	}
	valid, path := verifyGamePath(selectedDir)
	runtime.EventsEmit(a.ctx, "search_success", path)
	Success("游戏路径有效：%s", selectedDir)
	return map[string]any{"path": selectedDir, "valid": valid}
}

// StartAutoSearch 开始自动搜索
func (a *App) StartAutoSearch() {
	go func() {
		defaultPath := GetDefaultWarThunderPath()
		isValid, path := verifyGamePath(defaultPath)
		if isValid {
			Info("在默认位置找到游戏：%s", path)
			runtime.EventsEmit(a.ctx, "search_success", path)
			return
		}

		Warn("默认路径未找到，正在全盘搜寻游戏文件夹...")
		foundPath := FindGameDir()

		if foundPath != "" {
			Success("找到游戏路径：%s", foundPath)
			runtime.EventsEmit(a.ctx, "search_success", foundPath)
		} else {
			Error("未能在您的电脑上找到 War Thunder 安装目录，请使用手动选择")
			runtime.EventsEmit(a.ctx, "search_fail")
		}
	}()
}

// ClearLogs 清空日志
func (a *App) ClearLogs() {
	return
}

// GetVoiceList 获取语音包库列表
func (a *App) GetVoiceList() []map[string]any {
	voicePath := a.resolvePath(VoiceFolder)
	modDirs, err := GetInstalledMods(voicePath)
	if err != nil {
		Error("获取语音包列表失败: %v", err)
		return []map[string]any{}
	}

	var result []map[string]any
	for _, modDir := range modDirs {
		modPath := filepath.Join(voicePath, modDir)
		modInfo := a.buildModInfo(modDir, modPath)
		if modInfo != nil {
			result = append(result, modInfo)
		}
	}

	return result
}

// buildModInfo 构建单个 mod 的信息
func (a *App) buildModInfo(modId, modPath string) map[string]any {
	// 尝试读取元数据文件
	meta := ReadModMetadata(modPath)

	// 计算文件夹大小
	sizeStr := FormatSize(CalculateDirSize(modPath))

	// 检测 capabilities
	capabilities := DetectModCapabilities(modPath)

	// 获取子文件夹列表（用于安装模态框）
	folders := GetModFolders(modPath)

	// 获取修改时间
	dateStr := GetModifyTimeString(modPath)

	// 获取封面图片 URL
	coverURL := GetModCoverURL(modPath, meta)

	// 构建 mod 信息
	modInfo := map[string]any{
		"id":            modId,
		"title":         GetStringOrDefault(meta["title"], modId),
		"author":        GetStringOrDefault(meta["author"], "未知作者"),
		"version":       GetStringOrDefault(meta["version"], "1.0"),
		"note":          GetStringOrDefault(meta["note"], ""),
		"size_str":      sizeStr,
		"capabilities":  capabilities,
		"language":      GetLanguage(meta),
		"cover_url":     coverURL,
		"date":          dateStr,
		"link_video":    GetStringOrDefault(meta["link_video"], ""),
		"link_wtlive":   GetStringOrDefault(meta["link_wtlive"], ""),
		"link_bilibili": GetStringOrDefault(meta["link_bilibili"], ""),
		"folders":       folders,
	}

	return modInfo
}

// resolvePath 根据 FolderType 获取实际路径字符串
func (a *App) resolvePath(fType FolderType) string {
	if fType == GameFolder {
		return gamePath
	}

	if path, ok := FolderPaths[fType]; ok {
		return string(path)
	}

	return string(PendingFolderPath)
}

// ImportSelectedZip 导入选中的 ZIP
func (a *App) ImportSelectedZip(_type string) {
	selectedZip, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择压缩包",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "ZIP 文件",
				Pattern:     "*.zip",
			},
			{
				DisplayName: "7z 文件",
				Pattern:     "*.7z",
			},
			{
				DisplayName: "rar 文件",
				Pattern:     "*.rar",
			},
		},
	})
	if err != nil || selectedZip == "" {
		Error("未选择压缩包")
		return
	}
	a.ImportZips([]string{selectedZip}, _type)
}

func (a *App) ImportZipsFromPending(_type string) {

}

// ImportZips 批量导入 ZIP
func (a *App) ImportZips(selectedZips []string, typeStr string) {
	fType := FolderType(typeStr)
	targetDir := a.resolvePath(fType)

	for _, path := range selectedZips {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if _, err := os.Stat(filepath.Join(targetDir, name)); err == nil {
			Error("导入取消：检测到同名文件夹 %s 已存在", name)
			runtime.EventsEmit(a.ctx, "error_tip", "导入取消", "目录下已有重复的文件夹")
			return
		}
	}

	RunUnzipQueue(UnzipTask{
		Paths:     selectedZips,
		TargetDir: targetDir,
		OnProgress: func(current, total int, filename string) {
			Scan("进度 (%d/%d): %s", current, total, filename)
		},
		OnLog: func(level, message string) {
			Log(level, message)
		},
		OnFinished: func() {
			Scan("所有任务处理完毕")
			runtime.EventsEmit(a.ctx, "ev_import_finished", true)
		},
	})
}

// OpenFolder 打开文件夹
func (a *App) OpenFolder(folderType string) {
	// TODO: 根据类型打开对应文件夹
	Info("打开 %s", folderType)
	OpenFolder(FolderType(folderType))
}

// DeleteMod 删除语音包
func (a *App) DeleteMod(modId string) bool {
	// TODO: 实现删除语音包逻辑
	logger.Printf("删除语音包功能待实现: %s", modId)
	return false
}

// CheckInstallConflicts 检查安装冲突
func (a *App) CheckInstallConflicts(modId string, selectionJson string) []map[string]any {
	// TODO: 实现冲突检查逻辑
	logger.Printf("检查安装冲突功能待实现: %s", modId)
	return []map[string]any{}
}

// InstallMod 安装语音包
func (a *App) InstallMod(modId string, selectionJson string) {
	// TODO: 实现安装语音包逻辑
	logger.Printf("安装语音包功能待实现: %s", modId)
}

// RestoreGame 还原游戏
func (a *App) RestoreGame() {
	// TODO: 实现还原游戏逻辑
	logger.Println("还原游戏功能待实现")
}

// CheckFirstRun 检查首次运行
func (a *App) CheckFirstRun() map[string]any {
	agreementVersion := vp.GetString("agreement_version")
	currentVersion := "1.1" // TODO: 从配置或常量获取当前版本
	if agreementVersion != currentVersion {
		return map[string]any{"status": true, "version": currentVersion}
	}
	return map[string]any{"status": false, "version": agreementVersion}
}

// AgreeToTerms 同意条款
func (a *App) AgreeToTerms(version string) {
	vp.Set("agreement_version", version)
	vp.WriteConfig()
	Info("已同意条款，版本: %s", version)
}
