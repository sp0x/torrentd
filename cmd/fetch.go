package main

import (
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/torrent"
	"github.com/spf13/cobra"
)

func init() {
	cmdFetchTorrents := &cobra.Command{
		Use:   "fetch",
		Short: "Fetches torrents. If no flags are given this command simply fetches the latest 10 pages of torrents.",
		Run:   fetchTorrents,
	}
	rootCmd.AddCommand(cmdFetchTorrents)
}

func fetchTorrents(cmd *cobra.Command, args []string) {
	client := indexer.NewTorrentHelper(&appConfig)
	_ = torrent.GetNewTorrents(client, nil)
}
