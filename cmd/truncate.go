package main

import (
	"github.com/sp0x/rutracker-rss/torrent"
	"github.com/spf13/cobra"
)

func init() {
	cmdTruncate := &cobra.Command{
		Use:     "truncate",
		Aliases: []string{"t"},
		Short:   "Truncates the torrent database",
		Run:     truncateTorrentDb,
	}
	rootCmd.AddCommand(cmdTruncate)
}

func truncateTorrentDb(cmd *cobra.Command, args []string) {
	storage := torrent.Storage{}
	storage.Truncate()
}
