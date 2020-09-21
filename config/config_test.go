package config

import (
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetCachePath(t *testing.T) {
}

func TestGetDefinitionDirs(t *testing.T) {
	home, _ := homedir.Dir()
	homeDefs := filepath.FromSlash(path.Join(home, ".torrentd", "definitions"))
	pwd, _ := os.Getwd()
	pwdDefs := filepath.FromSlash(path.Join(pwd, "definitions"))
	tests := []struct {
		name string
		want []string
	}{
		{"", []string{pwdDefs, homeDefs}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDefinitionDirs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefinitionDirs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {

}
