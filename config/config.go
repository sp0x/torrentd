package config

import (
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
	dir := homeDirectory(".cache", appname, subdir)
	return dir
}

func SetDefaults(cfg Config) {
	_ = cfg.Set("definition.dirs", GetDefinitionDirs())
}

func GetDefinitionDirs() []string {
	dirs := []string{}

	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, "definitions"))
	}

	dirs = append(dirs, homeDirectory(".config", appname, "definitions"))

	if configDir := os.Getenv("CONFIG_DIR"); configDir != "" {
		dirs = append(dirs, filepath.Join(configDir, "definitions"))
	}

	return dirs
}

func homeDirectory(subdir ...string) string {
	dirs := []string{"~"}
	dirs = append(dirs, subdir...)
	dir := path.Join(dirs...)
	return dir
}
