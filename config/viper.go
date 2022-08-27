package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ViperConfig struct{}

func (v *ViperConfig) SetSiteOption(section, key, value string) error {
	viper.Set(fmt.Sprintf("indexers.%s.%s", section, key), value)
	return nil
}

func (v *ViperConfig) Set(key, value interface{}) {
	viper.Set(fmt.Sprintf("%s", key), value)
}

func (v *ViperConfig) GetSiteOption(name, key string) (string, bool, error) {
	indexerMap := viper.GetStringMap(fmt.Sprintf("indexers.%s", name))
	a, b := indexerMap[key]
	if !b {
		return "", b, nil
	}
	return a.(string), b, nil
}

func (v *ViperConfig) GetSite(name string) (map[string]string, error) {
	indexerMap := viper.GetStringMapString(fmt.Sprintf("indexers.%s", name))
	return indexerMap, nil
}

func (v *ViperConfig) Get(key string) interface{} {
	return viper.Get(key)
}

func (v *ViperConfig) GetInt(param string) int {
	return viper.GetInt(param)
}

func (v *ViperConfig) GetBool(param string) bool {
	return viper.GetBool(param)
}

func (v *ViperConfig) GetString(param string) string {
	return viper.GetString(param)
}

func (v *ViperConfig) GetBytes(param string) []byte {
	return []byte(viper.GetString(param))
}
