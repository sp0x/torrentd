package main

import (
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/sp0x/torrentd/config"
)

var appConfig config.ViperConfig

func initConfig() {
	// We load the default config file
	homeDir, _ := homedir.Dir()
	defaultConfigPath := ""
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		defaultConfigPath = path.Join(homeDir, ".torrentd")
		_ = os.MkdirAll(defaultConfigPath, os.ModePerm)
		viper.AddConfigPath(defaultConfigPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName("torrentd")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			config.SetDefaults(&appConfig)
			log.WithFields(log.Fields{
				"path": defaultConfigPath,
			}).Info("no config file found, creating one using defaults")
			err = viper.SafeWriteConfig()
			if err != nil {
				log.Warningf("error while writing default config file: %v\n", err)
			}
		} else {
			log.Warningf("error while reading config file: %v\n", err)
			os.Exit(1)
		}
	}

	log.SetLevel(config.GetMinLogLevel(&appConfig))
}
