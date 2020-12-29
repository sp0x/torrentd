package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
)

func init() {
	cmdTruncate := &cobra.Command{
		Use:     "truncate",
		Aliases: []string{"t"},
		Short:   "Truncates the database",
		Run:     truncateTorrentDB,
	}
	rootCmd.AddCommand(cmdTruncate)
}

func truncateTorrentDB(cmd *cobra.Command, args []string) {
	store := storage.NewBuilder().
		WithRecord(&search.ScrapeResultItem{}).
		Build()
	defer store.Close()
	fmt.Printf("not supported\n")
	os.Exit(1)
	_ = store.Truncate()
}
