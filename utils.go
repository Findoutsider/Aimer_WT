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
