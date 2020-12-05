package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

//var subtitleIndexer string

func init() {
	cmdFetchTorrents := &cobra.Command{
		Use:   "subs",
		Short: "Finds subtitles using indexers",
		Run:   findSubtitles,
	}
	//flags := cmdFetchTorrents.Flags()
	//flags.StringVarP(&subtitleIndexer, "indexer", "x", "subsunacs", "The subtitle site to use.")
	//_ = viper.BindPFlag("indexer", flags.Lookup("indexer"))
	//_ = viper.BindEnv("indexer")
	rootCmd.AddCommand(cmdFetchTorrents)
}

func findSubtitles(_ *cobra.Command, args []string) {
	helper := indexer.NewFacadeFromConfiguration(&appConfig)
	if helper == nil {
		os.Exit(1)
	}
	var currentSearch search.Instance
	var err error
	var searchQuery = strings.Join(args, " ")
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
