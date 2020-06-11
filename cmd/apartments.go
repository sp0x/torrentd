package main

import (
	"fmt"
	"github.com/sp0x/rutracker-rss/cmd/bots"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/categories"
	"github.com/sp0x/rutracker-rss/torznab"
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
	//flags := cmdGetApartments.Flags()

	_ = viper.BindEnv("indexer")
	_ = viper.BindEnv("telegram_token")
	rootCmd.AddCommand(cmdGetApartments)
}

func findAppartments(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	flags.StringVarP(&aptIndexer, "indexer", "x", "cityapartment", "The appartment site to use.")
	_ = viper.BindPFlag("indexer", flags.Lookup("indexer"))

	helper := indexer.NewAggregateIndexerHelperWithCategories(&appConfig, categories.Rental)
	if helper == nil {
		os.Exit(1)
	}
	var searchQuery = strings.Join(args, " ")
	interval := 30
	//Create our query
	query := torznab.ParseQueryString(searchQuery)
	query.Categories = []int{categories.Rental.ID}
	resultsChan := indexer.Watch(helper, query, interval)
	//Change this.
	chatMessagesChannel := make(chan bots.ChatMessage)
	telegram := bots.NewTelegram()
	go telegram.Run()
	go telegram.FeedBroadcast(chatMessagesChannel)
	for true {
		select {
		case result := <-resultsChan:
			//Log it if it's needed
			//
			if result.IsNew() || result.IsUpdate() {
				price := result.GetField("price")
				reserved := result.GetField("reserved")
				if reserved == "true" {
					reserved = "It's currently reserved"
				} else {
					reserved = "And not reserved yet!!!"
				}
				msgText := fmt.Sprintf("I found a new property\n"+
					"[%s](%s)\n"+
					"*%s* - %s", result.Title, result.Link, price, reserved)
				message := bots.ChatMessage{Text: msgText, Banner: result.Banner}
				chatMessagesChannel <- message
				area := result.Size
				fmt.Printf("[%s][%d][%s] %s - %s\n", price, area, reserved, result.ResultItem.Title, result.Link)
			}
		}
	}

	//We store them here also, so we have faster access
	//bolts := storage.BoltStorage{}
	//_ = bolts.StoreSearchResults(currentSearch.GetResults())
	//for _, r := range currentSearch.GetResults() {
	//
	//}
}
