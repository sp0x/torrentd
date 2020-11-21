package main

import (
	"fmt"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

func init() {
	cmdGetInfo := &cobra.Command{
		Use:     "info",
		Aliases: []string{"inspect"},
		Short:   "Lists torrent DB info.",
		Run:     getInfo,
	}
	rootCmd.AddCommand(cmdGetInfo)
}

func getInfo(cmd *cobra.Command, args []string) {
	store := storage.NewBuilder().
		WithRecord(&search.ExternalResultItem{}).
		Build()
	defer store.Close()

	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	stats := store.GetStats()
	if stats == nil {
		fmt.Print("No stats information found.")
		os.Exit(1)
	}
	for _, namespace := range stats.Namespaces {
		_, _ = fmt.Fprintf(tabWr, "[%d]\t%s\n", namespace.RecordCount, namespace.Name)
	}
	_ = tabWr.Flush()
}
