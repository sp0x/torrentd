package config

import (
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
	"path/filepath"
)

var appname = "tracker-rss"

type Config interface {
	GetSiteOption(name, key string) (string, bool, error)
	GetSite(section string) (map[string]string, error)
	GetInt(param string) int
	GetString(s string) string
	GetBytes(s string) []byte
	SetSiteOption(section, key, value string) error
	Set(key, value interface{}) error
}

func GetCachePath(subdir string) string {
	home, _ := homedir.Dir()
	homeDefsDir := path.Join(home, "."+appname, "cache", subdir)
	_ = os.MkdirAll(homeDefsDir, os.ModePerm)
	return homeDefsDir
}

func SetDefaults(cfg Config) {
	_ = cfg.Set("definition.dirs", GetDefinitionDirs())
}

func init() {
	home, _ := homedir.Dir()
	homeDefsDir := path.Join(home, "."+appname, "definitions")
	_ = os.MkdirAll(homeDefsDir, os.ModePerm)
}

func GetDefinitionDirs() []string {
	dirs := []string{}
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, "definitions"))
	}

	home, _ := homedir.Dir()
	homeDefsDir := path.Join(home, "."+appname, "definitions")

	dirs = append(dirs, homeDefsDir)
	if configDir := os.Getenv("CONFIG_DIR"); configDir != "" {
		dirs = append(dirs, filepath.Join(configDir, "definitions"))
	}

	return dirs
}
