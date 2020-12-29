package main

import (
	"github.com/spf13/cobra"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/torrent"
)

func init() {
	cmdFetchTorrents := &cobra.Command{
		Use:   "fetch",
		Short: "Fetches torrents. If no flags are given this command simply fetches the latest 10 pages of torrents.",
		Run:   fetchTorrents,
	}
	rootCmd.AddCommand(cmdFetchTorrents)
}

func fetchTorrents(_ *cobra.Command, _ []string) {
	facade := indexer.NewFacadeFromConfiguration(&appConfig)
	_ = torrent.GetNewScrapeItems(facade, nil)
}
