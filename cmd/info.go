package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
)

var dumpAdditionalInfo = false

func init() {
	cmdGetInfo := &cobra.Command{
		Use:     "info",
		Aliases: []string{"inspect"},
		Short:   "Lists torrent DB info.",
		Run:     getInfo,
	}
	cmdGetInfo.Flags().BoolVarP(&dumpAdditionalInfo, "dump", "d", false, "Dump additional info")
	rootCmd.AddCommand(cmdGetInfo)
}

func getInfo(_ *cobra.Command, _ []string) {
	store := storage.NewBuilder(&appConfig).
		WithRecord(&search.ScrapeResultItem{}).
		Build()
	defer store.Close()

	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	stats := store.GetStats(dumpAdditionalInfo)
	if stats == nil {
		fmt.Print("No stats information found.")
		return
	}
	for _, namespace := range stats.Namespaces {
		_, _ = fmt.Fprintf(tabWr, "[%d]\t%s\n", namespace.RecordCount, namespace.Name)
	}
	_ = tabWr.Flush()
}
