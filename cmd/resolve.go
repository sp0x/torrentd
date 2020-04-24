package main

import (
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
	client := torrent.NewTorrentHelper(&appConfig)
	torrent.ResolveTorrents(client, resolutionHours)
}
