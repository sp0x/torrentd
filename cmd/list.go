package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
)

var torrentCount = 500

func init() {
	cmdTruncate := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Lists the latest 100 torrents in the database",
		Run:     listLatestTorrents,
	}
	cmdTruncate.Flags().IntVarP(&torrentCount, "count", "c", 500, "Number of torrents to display")
	rootCmd.AddCommand(cmdTruncate)
}

func listLatestTorrents(cmd *cobra.Command, args []string) {
	store := storage.NewBuilder().
		WithRecord(&search.ScrapeResultItem{}).
		Build()
	defer store.Close()
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	torrents := store.GetLatest(torrentCount)
	for _, tr := range torrents {
		_, _ = fmt.Fprintf(tabWr, "%s:\t%s", tr.UUID(), tr.String())
		_ = tabWr.Flush()
	}
}
