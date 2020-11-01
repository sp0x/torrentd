package main

import (
	"fmt"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	cmdTruncate := &cobra.Command{
		Use:     "truncate",
		Aliases: []string{"t"},
		Short:   "Truncates the database",
		Run:     truncateTorrentDb,
	}
	rootCmd.AddCommand(cmdTruncate)
}

func truncateTorrentDb(cmd *cobra.Command, args []string) {
	store := storage.NewBuilder().
		WithRecord(&search.ExternalResultItem{}).
		Build()
	defer store.Close()
	fmt.Printf("not supported\n")
	os.Exit(1)
	//store.Truncate()
	//storage.Truncate()
}
