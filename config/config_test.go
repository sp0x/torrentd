package config

import (
	"github.com/mitchellh/go-homedir"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestGetCachePath(t *testing.T) {
	type args struct {
		subdir string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCachePath(tt.args.subdir); got != tt.want {
				t.Errorf("GetCachePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefinitionDirs(t *testing.T) {
	home, _ := homedir.Dir()
	homeDefs := path.Join(home, ".torrentd", "definitions")
	pwd, _ := os.Getwd()
	pwdDefs := path.Join(pwd, "definitions")
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
	type args struct {
		cfg Config
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//SetDefaults(cfg)
		})
	}
}