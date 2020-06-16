package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/server"
	"github.com/sp0x/torrentd/torrent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var watchInterval int

func init() {
	cmdWatch := &cobra.Command{
		Use:   "watch",
		Short: "Watches the torrent tracker for new torrents.",
		Run:   watchTracker,
	}
	cmdWatch.Flags().IntVarP(&watchInterval, "interval", "i", 10, "Interval between checks.")
	viper.SetDefault("port", 5000)
	_ = viper.BindEnv("port")
	_ = viper.BindEnv("api_key")
	rootCmd.AddCommand(cmdWatch)
}

func watchTracker(_ *cobra.Command, _ []string) {
	helper := indexer.NewFacadeFromConfiguration(&appConfig)
	if helper == nil {
		log.Error("Couldn't initialize torrent helper.")
		return
	}
	//Init the server
	go func() {
		rserver := server.NewServer(&appConfig)
		err := rserver.Listen(helper)
		if err != nil {
			fmt.Print(err)
		}
	}()

	//Start watching the torrent tracker.
	torrent.Watch(helper, watchInterval)
}
