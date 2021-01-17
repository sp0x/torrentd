package bots

import (
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/storage/indexing"
)

type TelegramRunner struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	storage storage.ItemStorage
}

type TelegramProvider func(token string) (*tgbotapi.BotAPI, error)

// NewTelegram creates a new telegram bot runner.
// This function uses the `chat_db` environment variable for storing the chats.
func NewTelegram(token string, cfg config.Config, provider TelegramProvider) (*TelegramRunner, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	if provider == nil {
		return nil, errors.New("telegram api provider is required")
	}
	telegram := &TelegramRunner{}
	bot, err := provider(token) // tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	telegram.bot = bot
	// bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	storageType := cfg.GetString("storage")
	if storageType == "" {
		storageType = "boltdb"
		// panic("no storage type configured")
	}
	telegram.storage = storage.NewBuilder().
		WithNamespace("__chats_telegram").
		WithPK(indexing.NewKey("id")).
		WithBacking(storageType).
		WithEndpoint(viper.GetString("chat_db")).
		WithRecord(&Chat{}).
		Build()
	// telegram.bolts = bolts
	return telegram, nil
}

// listenForUpdates listens from the telegram api.
func (t *TelegramRunner) listenForUpdates() {
	// Listen for people connecting to us
	go func() {
		for update := range t.updates {
			if update.Message == nil { // ignore any non-Message Updates
				continue
			}
			// We create our chat
			_ = t.storage.Add(&Chat{
				Username:    update.Message.From.UserName,
				InitialText: update.Message.Text,
				ChatID:      update.Message.Chat.ID,
			})
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			// reply := update.Message.Text
			if update.Message.Text == "/start" {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello. I'll keep you posted for new apartments.")
				// msg.ReplyToMessageID = update.Message.MessageID
				_, _ = t.bot.Send(msg)
			}
		}
	}()
}

// Run the bot, listening for updates from users
func (t *TelegramRunner) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := t.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}
	t.updates = updates
	t.listenForUpdates()
	return nil
}

// ForEachChat goes over all the persisted chats and invokes the callback on them.
func (t *TelegramRunner) ForEachChat(callback func(chat search.Record)) {
	t.storage.ForEach(callback)
}

// Broadcast a message to all active chats.
func (t *TelegramRunner) Broadcast(message *ChatMessage) {
	t.ForEachChat(func(obj search.Record) {
		chat := obj.(*Chat)
		msg := tgbotapi.NewMessage(chat.ChatID, message.Text)
		msg.DisableWebPagePreview = false
		msg.ParseMode = "markdown"
		// Since we're not replying.
		// msg.ReplyToMessageID = update.Message.MessageID
		_, _ = t.bot.Send(msg)
		if message.Banner != "" {
			imgMsg := tgbotapi.NewPhotoUpload(chat.ChatID, nil)
			imgMsg.FileID = message.Banner
			imgMsg.UseExisting = true
			_, _ = t.bot.Send(imgMsg)
		}
	})
}

// FeedBroadcast the messages that are passed to each one of the chats.
func (t *TelegramRunner) FeedBroadcast(messageChannel <-chan ChatMessage) error {
	if messageChannel == nil {
		return fmt.Errorf("message channel is required")
	}
	for chatMsg := range messageChannel {
		tmpChatMsg := chatMsg
		t.Broadcast(&tmpChatMsg)
	}
	return nil
}
