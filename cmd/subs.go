package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var subtitleIndexer string

func init() {
	cmdFetchTorrents := &cobra.Command{
		Use:   "subs",
		Short: "Finds subtitles using indexers",
		Run:   findSubtitles,
	}
	flags := cmdFetchTorrents.Flags()
	flags.StringVarP(&subtitleIndexer, "indexer", "x", "subsunacs", "The subtitle site to use.")
	_ = viper.BindPFlag("indexer", flags.Lookup("indexer"))
	_ = viper.BindEnv("indexer")
	rootCmd.AddCommand(cmdFetchTorrents)
}

func findSubtitles(cmd *cobra.Command, args []string) {
	helper := indexer.NewIndexerHelper(&appConfig)
	if helper == nil {
		os.Exit(1)
	}
	var currentSearch search.Instance
	var err error
	var searchQuery = strings.Join(args, " ")
	currentSearch, err = helper.Search(nil, searchQuery, 0)
	if err != nil {
		log.Error("Couldn't search for subtitles.")
		os.Exit(1)
	}
	for _, r := range currentSearch.GetResults() {
		print(r.ResultItem.Title)
	}
}
