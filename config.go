package main

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func InitConfig() {
	vp = viper.New()
	vp.SetConfigName("config")
	vp.SetConfigType("yaml")
	configDir := filepath.Join(basePath, "conf")
	vp.AddConfigPath(configDir)

	vp.SetDefault("game_path", "")
	vp.SetDefault("theme_mode", "Light")
	vp.SetDefault("is_first_run", true)
	vp.SetDefault("agreement_version", "")
	vp.SetDefault("active_theme", "default.json")
	vp.SetDefault("current_mod", "")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}
	if err := vp.ReadInConfig(); err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			configPath := filepath.Join(configDir, "config.yaml")
			err = vp.WriteConfigAs(configPath)
			if err != nil {
				logger.Println("创建配置文件失败:", err)
			} else {
				logger.Println("成功创建初始化配置文件")
			}
		}
	}
}

func GetActiveTheme() string {
	activeTheme := vp.GetString("active_theme")
	if activeTheme == "" {
		return "default.json"
	}
	return activeTheme
}
