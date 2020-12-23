package main

import (
	"fmt"
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
	cobra.OnInitialize(initConfig)
	flags := rootCmd.PersistentFlags()
	var verbose bool
	var dumpInputData bool
	index := ""
	flags.BoolVarP(&verbose, "verbose", "v", false, "Show more logs.")
	flags.BoolVarP(&dumpInputData, "dump", "", false, "Dump input data.")
	flags.StringVar(&configFile, "config", "", "The configuration file to use. By default it is ~/.torrentd/.tracker-rss.yaml")
	flags.StringVarP(&index, "index", "x", "", "The index to use. If you need to use multiple you can separate them with a comma.")
	_ = viper.BindPFlag("verbose", flags.Lookup("verbose"))
	_ = viper.BindEnv("verbose")

	_ = viper.BindPFlag("dump", flags.Lookup("dump"))
	_ = viper.BindEnv("dump")

	_ = viper.BindPFlag("index", flags.Lookup("index"))
	_ = viper.BindEnv("index")
	viper.SetEnvPrefix("TRACKER")
}

func main() {
	_ = os.MkdirAll("./db", os.ModePerm)
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
