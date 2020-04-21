package main

import (
	"fmt"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "rutracker",
	Short: "Gathers torrents from rutracker and serves them through a RSS server.",
}

func init() {
	//Init our db
	_ = os.MkdirAll("./db", os.ModePerm)
	gormDb := db.GetOrmDb()
	defer gormDb.Close()
	gormDb.AutoMigrate(&search.ExternalResultItem{})
	cobra.OnInitialize(initConfig)
	flags := rootCmd.PersistentFlags()
	var username, password string
	var verbose bool
	flags.StringVarP(&username, "username", "u", "", "The username to use")
	flags.StringVarP(&password, "password", "p", "", "The password to use")
	flags.BoolVarP(&verbose, "verbose", "v", false, "Show more logs")
	_ = viper.BindPFlag("username", flags.Lookup("username"))
	_ = viper.BindPFlag("password", flags.Lookup("password"))
	_ = viper.BindPFlag("verbose", flags.Lookup("verbose"))
	_ = viper.BindEnv("verbose")
	viper.SetEnvPrefix("TRACKER")
	_ = viper.BindEnv("username")
	_ = viper.BindEnv("password")

}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
