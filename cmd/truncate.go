package main

import (
	"fmt"
	"github.com/spf13/cobra"
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

func truncateTorrentDB(_ *cobra.Command, _ []string) {
	fmt.Printf("not supported\n")
	//store := storage.NewBuilder().
	//	WithRecord(&search.ScrapeResultItem{}).
	//	Build()
	//defer store.Close()
	//_ = store.Truncate()
}
