package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
)

// var subtitleIndexer string

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
	var currentSearch search.Instance
	var err error
	searchQuery := strings.Join(args, " ")
	subCat := categories.Subtitle
	currentSearch, err = helper.SearchKeywordsWithCategory(nil, searchQuery, subCat, 0)
	if err != nil {
		log.Error("Couldn't search for subtitles.")
		os.Exit(1)
	}
	for _, r := range currentSearch.GetResults() {
		scrape := r.AsScrapeItem()
		fmt.Printf("%s - %s\n", r.String(), scrape.Link)
	}
}
