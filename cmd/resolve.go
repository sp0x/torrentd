package main

import (
	"fmt"
	"os"

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
	facade, err := indexer.NewFacadeFromConfiguration(&appConfig)
	if err != nil {
		fmt.Printf("Couldn't initialize facade: %s", err)
		os.Exit(1)
	}
	indexes := facade.Indexes
	torrent.ResolveTorrents(indexes, &appConfig)
}
