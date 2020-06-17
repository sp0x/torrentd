package bots

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	. "github.com/onsi/gomega"
	"github.com/sp0x/torrentd/storage"
	"reflect"
	"testing"
)

//A mock for a telegram provider.
func MockedApiProvider(string) (*tgbotapi.BotAPI, error) {
	api := &tgbotapi.BotAPI{}
	return api, nil
}

func TestNewTelegram(t *testing.T) {
	g := NewGomegaWithT(t)
	type args struct {
		token    string
		provider TelegramProvider
	}
	tests := []struct {
		name string
		args *args
		want *TelegramRunner
	}{
		{"should return null on empty token", &args{"", nil}, nil},
		{"should return null on empty token", &args{"asd", nil}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := NewTelegram(tt.args.token, tt.args.provider); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTelegram() = %v, want %v", got, tt.want)
			}
		})
	}
	tgram, err := NewTelegram("asd", tgbotapi.NewBotAPI)
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(tgram).Should(BeNil())
	tgram, err = NewTelegram("asd", MockedApiProvider)
	g.Expect(err).Should(BeNil())
	g.Expect(tgram).ShouldNot(BeNil())

	g.Expect(tgram.bolts).ShouldNot(BeNil())
	g.Expect(tgram.updates).Should(BeNil())
	g.Expect(tgram.bot).ShouldNot(BeNil())
}

func TestTelegramRunner_FeedBroadcast(t *testing.T) {
	g := NewGomegaWithT(t)
	type fields struct {
		bot     *tgbotapi.BotAPI
		updates tgbotapi.UpdatesChannel
		bolts   *storage.BoltStorage
	}
	type args struct {
		messageChannel chan ChatMessage
	}
	defFields := fields{
		bot:     nil,
		updates: nil,
		bolts:   nil,
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   interface{}
	}{
		{"should return an error if a nil channel is passed", defFields, args{nil}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			tgram, _ := NewTelegram("asd", MockedApiProvider)
			//Do the feeding
			go func() {
				err := tgram.FeedBroadcast(tt.args.messageChannel)
				g.Expect(err).ShouldNot(BeNil())
			}()
			//tt.args.messageChannel <- ChatMessage{
			//	Text:   "",
			//	Banner: "",
			//}
			//close(tt.args.messageChannel)
		})
	}
}

func TestTelegramRunner_ForEachChat(t *testing.T) {
	g := NewGomegaWithT(t)
	tgram, _ := NewTelegram("asd", MockedApiProvider)
	chats := 0
	//Here we need to push a chat to telegram's bolt chat storage
	tgram.ForEachChat(func(chat *storage.Chat) {
		chats++
	})
	g.Expect(chats).ShouldNot(Equal(0))

}

func TestTelegramRunner_Run(t *testing.T) {
	g := NewGomegaWithT(t)
	tgram, _ := NewTelegram("asd", MockedApiProvider)
	//Here we need to run the bot and verify that it indeed receives a message
	g.Expect(tgram).ShouldNot(BeNil())
}

func TestTelegramRunner_listenForUpdates(t1 *testing.T) {
	type fields struct {
		bot     *tgbotapi.BotAPI
		updates tgbotapi.UpdatesChannel
		bolts   *storage.BoltStorage
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			//t := &TelegramRunner{
			//	bot:     tt.fields.bot,
			//	updates: tt.fields.updates,
			//	bolts:   tt.fields.bolts,
			//}
		})
	}
}
