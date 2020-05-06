package main

import (
	"fmt"
	storage "github.com/sp0x/rutracker-rss/storage"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

var torrentCount int = 500

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
	st := storage.DBStorage{}
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	torrents := st.GetLatest(torrentCount)
	for _, tr := range torrents {
		_, _ = fmt.Fprintf(tabWr, "%s\t%s\t%s", tr.LocalCategoryID, tr.Title, tr.AddedOnStr())
		_ = tabWr.Flush()
	}
}
