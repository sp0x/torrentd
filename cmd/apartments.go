package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/torznab"
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
	query.Page = 0
	query.Categories = []int{categories.Rental.ID}
	resultsChan := indexer.Watch(helper, query, interval)
	//Change this.
	chatBroadcastChan := make(chan search.ExternalResultItem)
	go runBot(chatBroadcastChan)
	for true {
		select {
		case result := <-resultsChan:
			//log.Infof("New result: %s\n", result)
			chatBroadcastChan <- result
			if result.IsNew() || result.IsUpdate() {
				price := result.GetField("price")
				reserved := result.GetField("reserved")
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

func runBot(itemsChannel <-chan search.ExternalResultItem) {
	token := viper.GetString("telegram_token")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	//bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	bolts, _ := storage.NewBoltStorage("")

	//Listen for people connecting to us
	go func() {
		for update := range updates {
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}
			//We create our chat
			bolts.StoreChat(&storage.Chat{
				Username:    update.Message.From.UserName,
				InitialText: update.Message.Text,
				ChatId:      update.Message.Chat.ID,
			})
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			//reply := update.Message.Text
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello. I'll keep you posted for new apartments.")
				//msg.ReplyToMessageID = update.Message.MessageID
				_, _ = bot.Send(msg)
			}
		}
	}()
	for item := range itemsChannel {
		if !item.IsNew() && !item.IsUpdate() {
			continue
		}
		price := item.GetField("price")
		reserved := item.GetField("reserved")
		if reserved == "true" {
			reserved = "It's currently reserved"
		} else {
			reserved = "And not reserved yet!!!"
		}
		_ = bolts.ForChat(func(chat *storage.Chat) {
			msgText := fmt.Sprintf("I found a new property\n"+
				"[%s](%s)\n"+
				"*%s* - %s", item.Title, item.Link, price, reserved)
			msg := tgbotapi.NewMessage(chat.ChatId, msgText)
			msg.DisableWebPagePreview = false
			msg.ParseMode = "markdown"
			//Since we're not replying.
			//msg.ReplyToMessageID = update.Message.MessageID
			_, _ = bot.Send(msg)
			imgMsg := tgbotapi.NewPhotoUpload(chat.ChatId, nil)
			imgMsg.FileID = item.Banner
			imgMsg.UseExisting = true
			_, _ = bot.Send(imgMsg)
		})
	}

}
