package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/config"
	"github.com/spf13/viper"
)

var appConfig config.ViperConfig

func initConfig() {
	//We load the default config file
	viper.AddConfigPath("~/.config/tracker-rss")
	viper.SetConfigType("yaml")
	viper.SetConfigName(".tracker-rss")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.SafeWriteConfig()
			if err != nil {
				log.Warningf("error while writing default config file: %v\n", err)
			}
		} else {
			log.Warningf("error while reading config file: %v\n", err)
		}
	}
	if viper.GetBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}
	config.SetDefaults(&appConfig)
}
