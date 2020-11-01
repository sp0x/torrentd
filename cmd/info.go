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

	//tc := store.Size()
	fmt.Printf("Not supported for now!\n")
	os.Exit(1)
	//tCategories := store.GetCategories()
	//fmt.Printf("Torrent count: %d\n", tc)
	//fmt.Printf("Categories: \n")
	//for _, tr := range tCategories {
	//	_, _ = fmt.Fprintf(tabWr, "[%s]\t%s\n", tr.CategoryId, tr.CategoryName)
	//	_ = tabWr.Flush()
	//}
}
