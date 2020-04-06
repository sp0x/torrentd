package main

import (
	"fmt"
	"github.com/sp0x/rutracker-rss/torrent"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	cmdFetchTorrents := &cobra.Command{
		Use:   "fetch",
		Short: "Fetches torrents. If no flags are given this command simply fetches the latest 10 pages of torrents.",
		Run:   fetchTorrents,
	}
	rootCmd.AddCommand(cmdFetchTorrents)
}

func fetchTorrents(cmd *cobra.Command, args []string) {
	client := torrent.NewRutracker()
	err := client.Login(viper.GetString("username"), viper.GetString("password"))
	if err != nil {
		fmt.Println("Could not login to tracker.")
		os.Exit(1)
	}
	_ = torrent.GetNewTorrents(client, nil)
}
