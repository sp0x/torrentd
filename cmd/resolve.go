package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/torrent"
	"github.com/spf13/cobra"
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

func resolveTorrents(cmd *cobra.Command, args []string) {
	helper := torrent.NewTorrentHelper(&appConfig)
	if helper == nil {
		log.Error("Couldn't initialize torrent helper.")
		return
	}
	torrent.ResolveTorrents(helper, resolutionHours)
}
