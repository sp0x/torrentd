package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/categories"
)

func init() {
	cmdFetchTorrents := &cobra.Command{
		Use:   "subs",
		Short: "Finds subtitles using indexers",
		Run:   findSubtitles,
	}
	rootCmd.AddCommand(cmdFetchTorrents)
}

func findSubtitles(_ *cobra.Command, args []string) {
	helper := indexer.NewFacadeFromConfiguration(&appConfig)
	if helper == nil {
		os.Exit(1)
	}
	searchQuery := strings.Join(args, " ")
	subCat := categories.Subtitle
	results, err := helper.SearchKeywordsWithCategory(searchQuery, 0, subCat)
	if err != nil {
		log.Error("Couldn't search for subtitles.")
		os.Exit(1)
	}
	for resultPage := range results {
		for _, r := range resultPage {
			scrape := r.AsScrapeItem()
			fmt.Printf("%s - %s\n", r.String(), scrape.Link)
		}
	}
}
