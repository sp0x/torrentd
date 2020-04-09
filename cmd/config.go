package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func initConfig() {
	//We load the default config file
	viper.AddConfigPath(".")
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
}
