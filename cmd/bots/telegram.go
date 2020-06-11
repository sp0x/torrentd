package bots

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/storage"
	"github.com/spf13/viper"
)

type ChatBotRunner interface {
	Run()
}

type TelegramRunner struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	bolts   *storage.BoltStorage
}

//NewTelegram creates a new telegram bot runner.
func NewTelegram() *TelegramRunner {
	telegram := &TelegramRunner{}
	token := viper.GetString("telegram_token")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	telegram.bot = bot
	//bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	telegram.updates = updates
	bolts, _ := storage.NewBoltStorage("")
	telegram.bolts = bolts
	return telegram
}

func (t *TelegramRunner) listenForUpdates() {
	//Listen for people connecting to us
	go func() {
		for update := range t.updates {
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}
			//We create our chat
			_ = t.bolts.StoreChat(&storage.Chat{
				Username:    update.Message.From.UserName,
				InitialText: update.Message.Text,
				ChatId:      update.Message.Chat.ID,
			})
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			//reply := update.Message.Text
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello. I'll keep you posted for new apartments.")
				//msg.ReplyToMessageID = update.Message.MessageID
				_, _ = t.bot.Send(msg)
			}
		}
	}()
}

func (t *TelegramRunner) Run() {
	t.listenForUpdates()
}

func (t *TelegramRunner) ForEachChat(callback func(chat *storage.Chat)) {
	_ = t.bolts.ForChat(callback)
}

type ChatMessage struct {
	Text   string
	Banner string
}

//FeedBroadcast the messages that are passed to each one of the chats.
func (t *TelegramRunner) FeedBroadcast(messageChannel chan ChatMessage) {
	for chatMsg := range messageChannel {
		t.ForEachChat(func(chat *storage.Chat) {
			msg := tgbotapi.NewMessage(chat.ChatId, chatMsg.Text)
			msg.DisableWebPagePreview = false
			msg.ParseMode = "markdown"
			//Since we're not replying.
			//msg.ReplyToMessageID = update.Message.MessageID
			_, _ = t.bot.Send(msg)
			imgMsg := tgbotapi.NewPhotoUpload(chat.ChatId, nil)
			imgMsg.FileID = chatMsg.Banner
			imgMsg.UseExisting = true
			_, _ = t.bot.Send(imgMsg)
		})
	}
}

func runBot(itemsChannel <-chan search.ExternalResultItem) {

}
