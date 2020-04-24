package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/server"
	"github.com/sp0x/rutracker-rss/torrent"
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

func watchTracker(cmd *cobra.Command, args []string) {
	client := torrent.NewTorrentHelper(&appConfig)
	if client == nil {
		log.Error("Couldn't initialize torrent helper.")
		return
	}

	go func() {
		rserver := server.NewServer(&appConfig)
		err := rserver.Listen(client)
		if err != nil {
			fmt.Print(err)
		}
	}()

	torrent.Watch(client, watchInterval)
}
