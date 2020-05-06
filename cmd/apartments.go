package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/categories"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var aptIndexer string

func init() {
	cmdGetApartments := &cobra.Command{
		Use:   "apartments",
		Short: "Finds appartments using indexers",
		Run:   findAppartments,
	}
	flags := cmdGetApartments.Flags()
	flags.StringVarP(&aptIndexer, "indexer", "x", "cityapartment", "The appartment site to use.")
	_ = viper.BindPFlag("indexer", flags.Lookup("indexer"))
	_ = viper.BindEnv("indexer")
	rootCmd.AddCommand(cmdGetApartments)
}

func findAppartments(cmd *cobra.Command, args []string) {
	helper := indexer.NewAggregateIndexerHelperWithCategories(&appConfig, categories.Rental)
	if helper == nil {
		os.Exit(1)
	}
	var currentSearch search.Instance
	var err error
	var searchQuery = strings.Join(args, " ")
	subCat := categories.Rental
	currentSearch, err = helper.SearchKeywordsWithCategory(nil, searchQuery, subCat, 0)
	if err != nil {
		log.Error("Couldn't search for subtitles.")
		os.Exit(1)
	}
	//We store them here also, so we have faster access
	bolts := storage.BoltStorage{}
	_ = bolts.StoreSearchResults(currentSearch.GetResults())
	for _, r := range currentSearch.GetResults() {
		fmt.Printf("%s - %s\n", r.ResultItem.Title, r.Link)
	}
}
