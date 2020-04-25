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
var configFile = ""

func init() {
	//Init our db
	_ = os.MkdirAll("./db", os.ModePerm)
	gormDb := db.GetOrmDb()
	defer gormDb.Close()
	gormDb.AutoMigrate(&search.ExternalResultItem{})
	cobra.OnInitialize(initConfig)
	flags := rootCmd.PersistentFlags()
	var verbose bool
	flags.BoolVarP(&verbose, "verbose", "v", false, "Show more logs")
	flags.StringVar(&configFile, "config", "", "The configuration file to use. By default it is ~/.tracker-rss/.tracker-rss.yaml")
	_ = viper.BindPFlag("verbose", flags.Lookup("verbose"))
	_ = viper.BindEnv("verbose")
	viper.SetEnvPrefix("TRACKER")

}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
