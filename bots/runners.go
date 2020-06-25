package bots

type ChatBotRunner interface {
	//Run Starts listening for messages from clients.
	Run()
	//FeedBroadcasts Broadcast anything that comes from a channel.
	FeedBroadcast(messageChannel <-chan ChatMessage)
}
