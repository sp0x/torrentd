package bots

import (
	"reflect"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	. "github.com/onsi/gomega"

	"github.com/sp0x/torrentd/config"
)

// A mock for a telegram provider.
func MockedAPIProvider(string) (*tgbotapi.BotAPI, error) {
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
	cfg := &config.ViperConfig{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := NewTelegram(tt.args.token, cfg, tt.args.provider); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTelegram() = %v, want %v", got, tt.want)
			}
		})
	}
	tgram, err := NewTelegram("asd", cfg, tgbotapi.NewBotAPI)
	g.Expect(err).ShouldNot(BeNil())
	g.Expect(tgram).Should(BeNil())
	tgram, err = NewTelegram("asd", cfg, MockedAPIProvider)
	g.Expect(err).Should(BeNil())
	g.Expect(tgram).ShouldNot(BeNil())

	//	g.Expect(tgram.bolts).ShouldNot(BeNil())
	g.Expect(tgram.updates).Should(BeNil())
	g.Expect(tgram.bot).ShouldNot(BeNil())
}
