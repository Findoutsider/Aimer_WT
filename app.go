package main

import (
	"context"
	"fmt"
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

	installedMods := getCurrentInstalledMods()

	return map[string]any{
		"game_path":      path,
		"path_valid":     isVerify,
		"theme":          theme,
		"active_theme":   GetActiveTheme(),
		"installed_mods": installedMods, // 返回所有已安装的 mod 列表
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
		runtime.EventsEmit(a.ctx, "ev_import_cancelled")
		return
	}
	a.ImportZips([]string{selectedZip}, _type)
}

func (a *App) ImportZipsFromPending() {
	path := GetPath(PendingFolder)
	Info(path)
	zipFromFolders := ReadZipFromFolders(path)
	a.ImportZips(zipFromFolders, "voice")
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
			progress := int(float64(current) / float64(total) * 100)
			runtime.EventsEmit(a.ctx, "ev_import_progress", progress, fmt.Sprintf("正在导入 %s (%d/%d)", filename, current, total))
			a.showInfoTip("导入中", fmt.Sprintf("正在导入 %s", filename), 3000)
		},
		OnLog: func(level, message string) {
			Log(level, message)
		},
		OnFinished: func() {
			Scan("所有任务处理完毕")
			runtime.EventsEmit(a.ctx, "ev_import_progress", 100, "导入完成")
			a.showInfoTip("导入完成", "导入完成", 3000)
			runtime.EventsEmit(a.ctx, "ev_import_finished", true)
			a.refreshVoice()
		},
	})
}

// OpenFolder 打开文件夹
func (a *App) OpenFolder(folderType string) {
	OpenFolder(GetPath(FolderType(folderType)))
}

// DeleteMod 删除语音包
func (a *App) DeleteMod(modId string) bool {
	err := os.RemoveAll(filepath.Join(GetPath(VoiceFolder), modId))
	if err != nil {
		Info("删除失败: %v", err)
		a.showErrorTip("删除失败", err.Error(), 3000)
		return false
	}
	Info("已删除 %s", modId)
	a.showInfoTip("删除成功", "已删除 "+modId, 3000)
	return true
}

// CheckInstallConflicts 检查安装冲突（只检查，不安装）
func (a *App) CheckInstallConflicts(modId string, selectionJson string) []map[string]any {
	gameVoicePath := GetPath(GameVoiceFolder)
	voicePath := GetPath(VoiceFolder)
	modPath := filepath.Join(voicePath, modId)
	manifestPath := filepath.Join(gameVoicePath, ".manifest.json")

	if err := ensureGameVoiceFolder(gameVoicePath); err != nil {
		return []map[string]any{
			{"file": "", "existing_mod": "", "new_mod": modId, "error": err.Error()},
		}
	}

	manifest, err := loadOrCreateManifest(manifestPath)
	if err != nil {
		return []map[string]any{
			{"file": "", "existing_mod": "", "new_mod": modId, "error": err.Error()},
		}
	}

	selectedFolders, err := parseSelectedFolders(selectionJson)
	if err != nil {
		return []map[string]any{
			{"file": "", "existing_mod": "", "new_mod": modId, "error": "解析选择列表失败"},
		}
	}

	filesToInstall := collectModFiles(modPath, selectedFolders)

	conflicts := checkFileConflicts(filesToInstall, manifest, modId)
	return conflicts
}

// InstallMod 安装语音包
func (a *App) InstallMod(modId string, selectionJson string) {
	gameVoicePath := GetPath(GameVoiceFolder)
	voicePath := GetPath(VoiceFolder)
	modPath := filepath.Join(voicePath, modId)
	manifestPath := filepath.Join(gameVoicePath, ".manifest.json")

	if err := ensureGameVoiceFolder(gameVoicePath); err != nil {
		Error("创建游戏语音文件夹失败: %v", err)
		a.showErrorTip("安装失败", "创建游戏语音文件夹失败", 5000)
		return
	}

	manifest, err := loadOrCreateManifest(manifestPath)
	if err != nil {
		Error("加载 manifest 失败: %v", err)
		a.showErrorTip("安装失败", "加载 manifest 失败", 5000)
		return
	}

	selectedFolders, err := parseSelectedFolders(selectionJson)
	if err != nil {
		Error("解析选择列表失败: %v", err)
		a.showErrorTip("安装失败", "解析选择列表失败", 5000)
		return
	}

	installedFiles, err := installModFiles(modPath, gameVoicePath, selectedFolders, manifest, modId)
	if err != nil {
		Error("安装文件失败: %v", err)
		a.showErrorTip("安装失败", err.Error(), 5000)
		return
	}

	if err := saveManifest(manifestPath, manifest, modId, installedFiles); err != nil {
		Error("保存 manifest 失败: %v", err)
		a.showErrorTip("安装失败", "保存 manifest 失败", 5000)
		return
	}

	// 确保 config.blk 中已开启 enable_mod:b=yes
	configPath := filepath.Join(gamePath, "config.blk")
	if err := ensureEnableModFlag(configPath, true); err != nil {
		Warn("更新 config.blk 失败: %v", err)
	}

	runtime.EventsEmit(a.ctx, "ev_install_success", modId)
	Success("成功安装 mod %s，共安装 %d 个文件", modId, len(installedFiles))
}

// RestoreGame 还原游戏
func (a *App) RestoreGame() {
	gameVoicePath := GetPath(GameVoiceFolder)
	manifestPath := filepath.Join(gameVoicePath, ".manifest.json")

	if PathExists(gameVoicePath) {
		entries, err := os.ReadDir(gameVoicePath)
		if err == nil {
			for _, entry := range entries {
				if entry.Name() == ".manifest.json" {
					continue
				}
				entryPath := filepath.Join(gameVoicePath, entry.Name())
				if err := os.RemoveAll(entryPath); err != nil {
					Warn("删除文件失败: %s, %v", entryPath, err)
				}
			}
		}
	}

	emptyManifest := Manifest{
		InstalledMods: make(map[string]ModInfo),
		FileMap:       make(map[string]string),
	}
	if err := WriteJSON(manifestPath, emptyManifest); err != nil {
		Error("清空 manifest.json 失败: %v", err)
	}

	configPath := filepath.Join(gamePath, "config.blk")
	if PathExists(configPath) {
		if err := ensureEnableModFlag(configPath, false); err != nil {
			Warn("更新 config.blk 失败: %v", err)
		}
	}

	runtime.EventsEmit(a.ctx, "ev_restore_success")
	Success("游戏已还原为纯净模式")
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

func (a *App) showInfoTip(title string, content string, duration uint) {
	if duration == 0 {
		duration = 5000
	}
	runtime.EventsEmit(a.ctx, "info_tip", title, content, duration)
}

func (a *App) showWarnTip(title string, content string, duration uint) {
	if duration == 0 {
		duration = 5000
	}
	runtime.EventsEmit(a.ctx, "warn_tip", title, content, duration)
}

func (a *App) showErrorTip(title string, content string, duration uint) {
	if duration == 0 {
		duration = 5000
	}
	runtime.EventsEmit(a.ctx, "error_tip", title, content, duration)
}

func (a *App) refreshVoice() {
	runtime.EventsEmit(a.ctx, "refresh_voice")
}
