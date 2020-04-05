package main

import "os"

func main() {
	getNewTorrents(os.Getenv("RSS_USER"), os.Getenv("RSS_PASS"))
}
