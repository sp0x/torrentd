package config

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

var appname = "torrentd"

//go:generate mockgen -destination=mocks/mock_config.go -package=mocks . Config
type Config interface {
	GetSiteOption(name, key string) (string, bool, error)
	GetSite(key string) (map[string]string, error)
	GetInt(key string) int
	GetString(key string) string
	GetBool(key string) bool
	GetBytes(key string) []byte
	Get(key string) interface{}
	SetSiteOption(section, key, value string) error
	Set(key, value interface{})
}

func GetMinLogLevel(c Config) log.Level {
	if c.GetBool("verbose") {
		return log.DebugLevel
	}
	return log.InfoLevel
}

func GetCachePath(subdir string) string {
	home, _ := homedir.Dir()
	homeDefsDir := path.Join(home, "."+appname, "cache", subdir)
	_ = os.MkdirAll(homeDefsDir, os.ModePerm)
	return homeDefsDir
}

func SetDefaults(cfg Config) {
	cfg.Set("definition.dirs", GetDefinitionDirs())
}

func init() {
	home, _ := homedir.Dir()
	homeDefsDir := path.Join(home, "."+appname, "definitions")
	_ = os.MkdirAll(homeDefsDir, os.ModePerm)
}

func GetDefinitionDirs() []string {
	var dirs []string
	if cwd, err := os.Getwd(); err == nil {
		dirs = append(dirs, filepath.Join(cwd, "definitions"))
	}
	home, _ := homedir.Dir()
	homeDefsDir := filepath.FromSlash(path.Join(home, "."+appname, "definitions"))

	dirs = append(dirs, homeDefsDir)
	if configDir := os.Getenv("CONFIG_DIR"); configDir != "" {
		configDir = filepath.FromSlash(configDir)
		dirs = append(dirs, filepath.Join(configDir, "definitions"))
	}

	return dirs
}
