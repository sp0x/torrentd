package main

import (
	"fmt"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "rutracker",
	Short: "Gathers torrents from rutracker and serves them through a RSS server.",
	Run:   root}

var username, password string

func init() {
	//Init our db
	gormDb := db.GetOrmDb()
	defer gormDb.Close()
	gormDb.AutoMigrate(&db.Torrent{})
	cobra.OnInitialize(initConfig)
	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&username, "username", "u", "", "The username to use")
	flags.StringVarP(&password, "password", "p", "", "The password to use")
	_ = viper.BindPFlag("username", flags.Lookup("username"))
	_ = viper.BindPFlag("password", flags.Lookup("password"))
}

func root(cmd *cobra.Command, args []string) {

}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
