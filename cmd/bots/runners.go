package bots

type ChatBotRunner interface {
	Run()
	FeedBroadcast(messageChannel <-chan ChatMessage)
}
