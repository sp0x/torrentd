package bots

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/storage"
)

type TelegramRunner struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	bolts   *storage.BoltStorage
}

type TelegramProvider func(token string) (*tgbotapi.BotAPI, error)

//NewTelegram creates a new telegram bot runner.
func NewTelegram(token string, provider TelegramProvider) (*TelegramRunner, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}
	if provider == nil {
		return nil, errors.New("telegram api provider is required")
	}
	telegram := &TelegramRunner{}
	bot, err := provider(token) //tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	telegram.bot = bot
	//bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	bolts, _ := storage.NewBoltStorage("")
	telegram.bolts = bolts
	return telegram, nil
}

//listenForUpdates listens from the telegram api.
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

//Run the bot, listening for updates from users
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

//ForEachChat goes over all the persisted chats and invokes the callback on them.
func (t *TelegramRunner) ForEachChat(callback func(chat *storage.Chat)) {
	_ = t.bolts.ForChat(callback)
}

type ChatMessage struct {
	Text   string
	Banner string
}

//FeedBroadcast the messages that are passed to each one of the chats.
func (t *TelegramRunner) FeedBroadcast(messageChannel <-chan ChatMessage) error {
	if messageChannel == nil {
		return fmt.Errorf("message channel is required")
	}
	for chatMsg := range messageChannel {
		t.ForEachChat(func(chat *storage.Chat) {
			msg := tgbotapi.NewMessage(chat.ChatId, chatMsg.Text)
			msg.DisableWebPagePreview = false
			msg.ParseMode = "markdown"
			//Since we're not replying.
			//msg.ReplyToMessageID = update.Message.MessageID
			_, _ = t.bot.Send(msg)
			if chatMsg.Banner != "" {
				imgMsg := tgbotapi.NewPhotoUpload(chat.ChatId, nil)
				imgMsg.FileID = chatMsg.Banner
				imgMsg.UseExisting = true
				_, _ = t.bot.Send(imgMsg)
			}
		})
	}
	return nil
}
