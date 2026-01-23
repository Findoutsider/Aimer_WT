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

	// 1. 读取原始字节流
	content, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}

	// 2. 将字节流解析为结构体
	err = json.Unmarshal(content, &data)
	if err != nil {
		return data, err
	}

	return data, nil
}

func WriteJSON(filePath string, data any) error {
	// 1. 将对象转为带缩进的 JSON 字节流 (Indent 为 4 个空格)
	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	// 2. 写入物理文件 (0644 是标准的读写权限)
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
	vp.WriteConfig()

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

func OpenFolder(folderType FolderType) error {
	var cmd *exec.Cmd
	path := string(getFolderPath(folderType))
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin": // macOS
		cmd = exec.Command("open", path)
	default: // Linux
		cmd = exec.Command("xdg-open", path)
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
			Error("创建文件夹失败 [%s]: %v", path, err)
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

	os.MkdirAll(realDest, 0755)

	for _, f := range r.File {
		fpath := filepath.Join(realDest, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(realDest)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
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
	var folders []map[string]any

	entries, err := os.ReadDir(modPath)
	if err != nil {
		return folders
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
			folders = append(folders, map[string]any{
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
		folders = append([]map[string]any{{
			"path": "根目录",
			"type": "folder",
		}}, folders...)
	}

	return folders
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
