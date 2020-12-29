package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/torrent"
)

var resolutionHours int

func init() {
	cmdResolve := &cobra.Command{
		Use:   "resolve",
		Short: "Goes through the torrent database resolving each torrent for name, size and trackers.",
		Run:   resolveTorrents,
	}
	cmdResolve.Flags().IntVarP(&resolutionHours, "hours", "t", 10, "Resolve only torrents that have been created at least a given amount of hours ago.")
	rootCmd.AddCommand(cmdResolve)
}

func resolveTorrents(_ *cobra.Command, _ []string) {
	helper := indexer.NewFacadeFromConfiguration(&appConfig)
	if helper == nil {
		log.Error("Couldn't initialize helper.")
		return
	}
	config := helper.Config
	index := helper.Indexer
	torrent.ResolveTorrents(index, config)
}
