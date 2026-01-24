package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func ReadJSON[T any](filePath string) (T, error) {
	var data T

	content, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}

	err = json.Unmarshal(content, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func WriteJSON(filePath string, data any) error {
	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, content, 0644)
}

func GetDefaultWarThunderPath() string {
	key, err := registry.OpenKey(windows.HKEY_CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()

	steamPath, _, err := key.GetStringValue("SteamPath")
	if err != nil {
		return ""
	}

	return filepath.Join(steamPath, "steamapps", "common", "War Thunder")
}

func FindGameDir() string {
	drivers := getLogicalDrives()

	// 常见的安装深度路径（建议缩短，提高命中率）
	commonSubPaths := []string{
		"SteamLibrary/steamapps/common",
		"Program Files (x86)/Steam/steamapps/common",
		"Program Files/Steam/steamapps/common",
		"Gaijin/Games",
		"Games",
	}

	for _, drive := range drivers {
		Scan("[DFS] 正在搜索磁盘 %s ...", drive)

		for _, sub := range commonSubPaths {
			fullPath := filepath.Join(drive, sub, "War Thunder")
			isValid, _ := verifyGamePath(fullPath)
			if isValid {
				return fullPath
			}
		}

		foundPath := ""
		filepath.WalkDir(drive, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				return nil
			}

			name := d.Name()
			if name == "Windows" || name == "ProgramData" || strings.HasPrefix(name, "$") || name == "AppData" {
				return filepath.SkipDir
			}

			if strings.EqualFold(name, "War Thunder") {
				isValid, _ := verifyGamePath(path)
				if isValid {
					foundPath = path
					return filepath.SkipAll
				}
			}

			if strings.Count(path, string(os.PathSeparator)) > 3 {
				return filepath.SkipDir
			}
			return nil
		})

		if foundPath != "" {
			return foundPath
		}
	}
	return ""
}
func verifyGamePath(path string) (bool, string) {
	if path == "" {
		return false, "路径为空"
	}

	if !PathExists(path) {
		return false, "指定路径不存在: " + path
	}

	if !PathExists(filepath.Join(path, "config.blk")) {
		return false, "指定路径下不存在 config.blk"
	}

	gamePath = path
	vp.Set("game_path", gamePath)
	err := vp.WriteConfig()
	if err != nil {
		Error("保存路径失败: %v", err)
	}

	return true, "校验通过"
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getLogicalDrives() []string {
	var drives []string
	if runtime.GOOS == "windows" {
		for _, drive := range "CDEFGHIJKLMNOPQRSTUVWXYZ" {
			d := string(drive) + ":\\"
			if _, err := os.Stat(d); err == nil {
				drives = append(drives, d)
			}
		}
	} else {
		drives = append(drives, "/")
	}
	return drives
}

func OpenFolder(folder string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", folder)
	case "darwin": // macOS
		cmd = exec.Command("open", folder)
	default: // Linux
		cmd = exec.Command("xdg-open", folder)
	}

	return cmd.Start()
}

func OpenAndSelect(filePath string) error {
	return exec.Command("explorer", "/select,", filePath).Start()
}

func InitAppFolders() {
	for _, path := range folders {
		Info(string(path))
		err := os.MkdirAll(string(path), 0755)
		if err != nil {
			Error("创建资源文件夹失败 [%s]: %v", path, err)
		}
	}
}

func getFolderPath(_type FolderType) FolderPath {
	return FolderPaths[_type]
}

func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	fileNameWithExt := filepath.Base(src)
	subdirName := strings.TrimSuffix(fileNameWithExt, filepath.Ext(fileNameWithExt))
	realDest := filepath.Join(dest, subdirName)

	err = os.MkdirAll(realDest, 0755)
	if err != nil {
		Error("创建文件夹失败 [%s]: %v", realDest, err)
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(realDest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(realDest)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				Error("创建文件夹失败 [%s]: %v", fpath, err)
				return err
			}
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func RunUnzipQueue(task UnzipTask) {
	if len(task.Paths) == 0 {
		if task.OnFinished != nil {
			task.OnFinished()
		}
		return
	}

	go func() {
		total := len(task.Paths)
		if len(task.Paths) == 1 {
			task.OnLog("INFO", fmt.Sprintf("开始导入任务 [%s]", task.Paths[0]))
		} else {
			task.OnLog("INFO", "开始批量导入任务")
		}

		for i, path := range task.Paths {
			filename := filepath.Base(path)

			if task.OnProgress != nil {
				task.OnProgress(i+1, total, filename)
			}

			err := Unzip(path, task.TargetDir)

			if err != nil {
				task.OnLog("ERROR", "解压失败 ["+filename+"]: "+err.Error())
			} else {
				task.OnLog("SUCCESS", "解压完成 ["+filename+"]")
			}
		}

		if task.OnFinished != nil {
			task.OnFinished()
		}
	}()
}

func GetInstalledMods(dirPath string) ([]string, error) {
	var mods []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			mods = append(mods, entry.Name())
		}
	}

	return mods, nil
}

// CalculateDirSize 计算文件夹大小
func CalculateDirSize(dirPath string) int64 {
	var size int64
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// FormatSize 格式化文件大小
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetModifyTime 获取文件夹修改时间
func GetModifyTime(dirPath string) time.Time {
	info, err := os.Stat(dirPath)
	if err != nil {
		return time.Now()
	}
	return info.ModTime()
}

// GetModifyTimeString 获取文件夹修改时间（格式化字符串）
func GetModifyTimeString(dirPath string) string {
	info, err := os.Stat(dirPath)
	if err != nil {
		return time.Now().Format("2006-01-02")
	}
	return info.ModTime().Format("2006-01-02")
}

// GetStringOrDefault 获取字符串或默认值
func GetStringOrDefault(value any, defaultValue string) string {
	if str, ok := value.(string); ok && str != "" {
		return str
	}
	return defaultValue
}

// GetLanguage 获取语言列表
func GetLanguage(meta map[string]any) []string {
	if lang, ok := meta["language"]; ok {
		if langList, ok := lang.([]any); ok {
			var result []string
			for _, l := range langList {
				if str, ok := l.(string); ok {
					result = append(result, str)
				}
			}
			if len(result) > 0 {
				return result
			}
		}
		if langStr, ok := lang.(string); ok && langStr != "" {
			return []string{langStr}
		}
	}
	return []string{"多语言"}
}

// ReadModMetadata 读取 mod 元数据文件（尝试多种文件名）
func ReadModMetadata(modPath string) map[string]any {
	metaFiles := []string{"mod.json", "info.json", "metadata.json", "modinfo.json"}
	for _, filename := range metaFiles {
		metaPath := filepath.Join(modPath, filename)
		if PathExists(metaPath) {
			meta, err := ReadJSON[map[string]any](metaPath)
			if err == nil {
				return meta
			}
		}
	}
	return make(map[string]any)
}

// GetModCoverURL 获取 mod 封面图片 URL
func GetModCoverURL(modPath string, meta map[string]any) string {
	if coverURL := GetStringOrDefault(meta["cover_url"], ""); coverURL != "" {
		return coverURL
	}

	// 尝试在 mod 文件夹中查找图片文件
	imageFiles := []string{
		"cover.png", "cover.jpg", "cover.jpeg",
		"card.png", "card.jpg", "card.jpeg",
		"preview.png", "preview.jpg", "preview.jpeg",
		"thumbnail.png", "thumbnail.jpg", "thumbnail.jpeg",
		"image.png", "image.jpg", "image.jpeg",
	}

	for _, imgFile := range imageFiles {
		imgPath := filepath.Join(modPath, imgFile)
		if PathExists(imgPath) {
			// 如果 mod 文件夹中有图片，返回空字符串
			// 前端会检测到空字符串并使用默认图片
			// 注意：如果需要显示 mod 文件夹中的图片，需要额外的后端 API 支持
			return ""
		}
	}

	return "assets/card_image.png"
}

// DetectModCapabilities 检测 mod 的功能
func DetectModCapabilities(modPath string) map[string]bool {
	caps := map[string]bool{
		"tank":   false,
		"air":    false,
		"naval":  false,
		"radio":  false,
		"status": false,
	}

	err := filepath.Walk(modPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		name := strings.ToLower(info.Name())
		relPath, _ := filepath.Rel(modPath, path)
		relPathLower := strings.ToLower(relPath)

		if strings.Contains(name, "tank") || strings.Contains(name, "ground") ||
			strings.Contains(relPathLower, "tank") || strings.Contains(relPathLower, "ground") {
			caps["tank"] = true
		}

		if strings.Contains(name, "air") || strings.Contains(name, "aircraft") ||
			strings.Contains(relPathLower, "air") || strings.Contains(relPathLower, "aircraft") {
			caps["air"] = true
		}

		if strings.Contains(name, "naval") || strings.Contains(name, "ship") ||
			strings.Contains(relPathLower, "naval") || strings.Contains(relPathLower, "ship") {
			caps["naval"] = true
		}

		if strings.Contains(name, "radio") || strings.Contains(relPathLower, "radio") {
			caps["radio"] = true
		}

		if strings.Contains(name, "status") || strings.Contains(name, "situation") ||
			strings.Contains(relPathLower, "status") || strings.Contains(relPathLower, "situation") {
			caps["status"] = true
		}

		return nil
	})

	if err != nil {
		// 如果扫描失败，设置默认值
		caps["tank"] = true
	}

	return caps
}

// GetModFolders 获取 mod 的子文件夹列表
func GetModFolders(modPath string) []map[string]any {
	var f []map[string]any

	entries, err := os.ReadDir(modPath)
	if err != nil {
		return f
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		subDirPath := filepath.Join(modPath, entry.Name())
		hasBankFile := false
		subEntries, _ := os.ReadDir(subDirPath)
		for _, subEntry := range subEntries {
			if !subEntry.IsDir() && strings.HasSuffix(strings.ToLower(subEntry.Name()), ".bank") {
				hasBankFile = true
				break
			}
		}

		if hasBankFile {
			folderType := DetectFolderType(entry.Name(), subDirPath)
			f = append(f, map[string]any{
				"path": entry.Name(),
				"type": folderType,
			})
		}
	}

	rootEntries, _ := os.ReadDir(modPath)
	hasRootBank := false
	for _, entry := range rootEntries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".bank") {
			hasRootBank = true
			break
		}
	}
	if hasRootBank {
		f = append([]map[string]any{{
			"path": "根目录",
			"type": "folder",
		}}, f...)
	}

	return f
}

// DetectFolderType 检测文件夹类型
func DetectFolderType(folderName, folderPath string) string {
	nameLower := strings.ToLower(folderName)
	pathLower := strings.ToLower(folderPath)

	if strings.Contains(nameLower, "ground") || strings.Contains(nameLower, "tank") ||
		strings.Contains(pathLower, "ground") || strings.Contains(pathLower, "tank") {
		return "ground"
	}
	if strings.Contains(nameLower, "air") || strings.Contains(nameLower, "aircraft") ||
		strings.Contains(pathLower, "air") || strings.Contains(pathLower, "aircraft") {
		return "aircraft"
	}
	if strings.Contains(nameLower, "radio") || strings.Contains(pathLower, "radio") {
		return "radio"
	}
	return "folder"
}

func ReadZipFromFolders(folderPath string) []string {
	var f []string
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return f
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".zip") {
			f = append(f, filepath.Join(folderPath, entry.Name()))
		}
	}
	return f
}

// ensureGameVoiceFolder 确保游戏语音文件夹存在
func ensureGameVoiceFolder(gameVoicePath string) error {
	if err := os.MkdirAll(gameVoicePath, 0755); err != nil {
		Error("创建游戏语音文件夹失败: %v", err)
		return err
	}
	return nil
}

// loadOrCreateManifest 加载或创建 manifest.json
func loadOrCreateManifest(manifestPath string) (*Manifest, error) {
	var manifest Manifest

	if PathExists(manifestPath) {
		manifestData, err := ReadJSON[Manifest](manifestPath)
		if err != nil {
			Warn("读取 manifest.json 失败，将创建新文件: %v", err)
			manifest = Manifest{
				InstalledMods: make(map[string]ModInfo),
				FileMap:       make(map[string]string),
			}
		} else {
			manifest = manifestData
			// 确保 map 已初始化
			if manifest.InstalledMods == nil {
				manifest.InstalledMods = make(map[string]ModInfo)
			}
			if manifest.FileMap == nil {
				manifest.FileMap = make(map[string]string)
			}
		}
	} else {
		// 创建空的 manifest
		manifest = Manifest{
			InstalledMods: make(map[string]ModInfo),
			FileMap:       make(map[string]string),
		}
		if err := WriteJSON(manifestPath, manifest); err != nil {
			Error("创建 manifest.json 失败: %v", err)
			return nil, err
		}
	}

	return &manifest, nil
}

// parseSelectedFolders 解析用户选择的文件夹列表
func parseSelectedFolders(selectionJson string) ([]string, error) {
	var selectedFolders []string
	if err := json.Unmarshal([]byte(selectionJson), &selectedFolders); err != nil {
		Error("解析选择列表失败: %v", err)
		return nil, err
	}
	return selectedFolders, nil
}

// collectModFiles 收集要安装的文件列表
func collectModFiles(modPath string, selectedFolders []string) []string {
	var filesToInstall []string

	for _, folder := range selectedFolders {
		var sourcePath string
		if folder == "根目录" {
			sourcePath = modPath
		} else {
			sourcePath = filepath.Join(modPath, folder)
		}

		// 遍历文件夹下的所有 .bank 文件
		err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".bank") {
				fileName := filepath.Base(path)
				filesToInstall = append(filesToInstall, fileName)
			}
			return nil
		})
		if err != nil {
			Warn("遍历文件夹 %s 失败: %v", folder, err)
		}
	}

	return filesToInstall
}

// checkFileConflicts 检查文件冲突
func checkFileConflicts(filesToInstall []string, manifest *Manifest, modId string) []map[string]any {
	var conflicts []map[string]any

	for _, fileName := range filesToInstall {
		if existingModId, exists := manifest.FileMap[fileName]; exists {
			if existingModId != modId {
				conflicts = append(conflicts, map[string]any{
					"file":         fileName,
					"existing_mod": existingModId,
					"new_mod":      modId,
				})
			}
			// 如果 existingModId == modId，说明是同一个 mod 重新安装
		}
	}

	return conflicts
}

// installModFiles 安装文件并更新 manifest
func installModFiles(modPath, gameVoicePath string, selectedFolders []string, manifest *Manifest, modId string) ([]string, error) {
	// 如果同一个 mod 重新安装，先清理旧的文件记录
	if oldModInfo, exists := manifest.InstalledMods[modId]; exists {
		// 从 file_map 中移除旧的文件记录
		for _, oldFile := range oldModInfo.Files {
			// 只有当 file_map 中该文件属于当前 mod 时才删除
			if manifest.FileMap[oldFile] == modId {
				delete(manifest.FileMap, oldFile)
			}
		}
	}

	var installedFiles []string

	for _, folder := range selectedFolders {
		var sourcePath string
		if folder == "根目录" {
			sourcePath = modPath
		} else {
			sourcePath = filepath.Join(modPath, folder)
		}

		err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".bank") {
				fileName := filepath.Base(path)
				destPath := filepath.Join(gameVoicePath, fileName)

				if err := copyFile(path, destPath); err != nil {
					Error("复制文件失败 %s -> %s: %v", path, destPath, err)
					return nil
				}

				installedFiles = append(installedFiles, fileName)
				manifest.FileMap[fileName] = modId
			}
			return nil
		})
		if err != nil {
			Warn("处理文件夹 %s 失败: %v", folder, err)
		}
	}

	return installedFiles, nil
}

// saveManifest 保存 manifest.json
func saveManifest(manifestPath string, manifest *Manifest, modId string, installedFiles []string) error {
	// 更新 installed_mods
	manifest.InstalledMods[modId] = ModInfo{
		Files:       installedFiles,
		InstallTime: time.Now().Format(time.RFC3339Nano),
	}

	// 保存 manifest.json
	if err := WriteJSON(manifestPath, manifest); err != nil {
		Error("保存 manifest.json 失败: %v", err)
		return err
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ensureEnableModFlag 确保 config.blk 中存在 enable_mod 标记
// enabled=true  -> enable_mod:b=yes
// enabled=false -> enable_mod:b=no
func ensureEnableModFlag(configPath string, enabled bool) error {
	if !PathExists(configPath) {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	text := string(data)
	targetValue := "yes"
	if !enabled {
		targetValue = "no"
	}

	lines := strings.Split(text, "\n")
	foundLine := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "enable_mod:b=") {
			// 保留原来的缩进
			prefix := line[:strings.Index(line, strings.TrimLeft(line, " \t"))]
			lines[i] = fmt.Sprintf("%senable_mod:b=%s", prefix, targetValue)
			foundLine = true
			break
		}
	}

	if foundLine {
		newText := strings.Join(lines, "\n")
		return os.WriteFile(configPath, []byte(newText), 0644)
	}

	// 没有找到 enable_mod 行
	// 尝试插入到 sound{ 块内，如果没有 sound{，则追加一个完整的 sound 块
	if idx := strings.Index(text, "sound{"); idx != -1 {
		insertPos := idx + len("sound{")
		insertLine := fmt.Sprintf("\n  enable_mod:b=%s", targetValue)
		newText := text[:insertPos] + insertLine + text[insertPos:]
		return os.WriteFile(configPath, []byte(newText), 0644)
	}

	// 没有 sound 块，追加一个完整的块
	block := fmt.Sprintf("\n\nsound{\n  fmod_sound_enable:b=yes\n  speakerMode:t=\"auto\"\n  enable_mod:b=%s\n}\n", targetValue)
	newText := text + block
	return os.WriteFile(configPath, []byte(newText), 0644)
}

// getCurrentInstalledMods 获取所有当前已安装的 mod（从 manifest 中获取）
func getCurrentInstalledMods() []string {
	gameVoicePath := GetPath(GameVoiceFolder)
	manifestPath := filepath.Join(gameVoicePath, ".manifest.json")

	if !PathExists(manifestPath) {
		return []string{}
	}

	manifest, err := ReadJSON[Manifest](manifestPath)
	if err != nil {
		return []string{}
	}

	// 返回所有已安装的 mod ID 列表
	var modIds []string
	for modId := range manifest.InstalledMods {
		modIds = append(modIds, modId)
	}

	return modIds
}
