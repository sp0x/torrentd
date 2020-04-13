package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type ViperConfig struct{}

func (v *ViperConfig) GetSiteOption(name, key string) (string, bool, error) {
	indexerMap := viper.GetStringMap(fmt.Sprintf("indexer.%s", name))
	a, b := indexerMap[key]
	return a.(string), b, nil
}

func (v *ViperConfig) GetSite(name string) (map[string]string, error) {
	indexerMap := viper.GetStringMapString(fmt.Sprintf("indexer.%s", name))
	return indexerMap, nil
}

func (v *ViperConfig) GetInt(param string) int {
	return viper.GetInt(param)
}

func (v *ViperConfig) GetString(param string) string {
	return viper.GetString(param)
}
