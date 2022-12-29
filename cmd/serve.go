package main

import (
	"fmt"
	"os"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cmdServe := &cobra.Command{
		Use:   "serve",
		Short: "Runs the RSS server.",
		Run:   serve,
	}
	storage := ""
	port := 5000
	cmdFlags := cmdServe.Flags()
	cmdFlags.StringVarP(&storage, "storage", "o", "boltdb", `The storage backing to use.
Currently supported storage backings: boltdb, firebase, sqlite`)
	cmdFlags.IntVarP(&port, "port", "p", 5000, "The port to listen on.")
	viper.SetDefault("port", 5000)
	_ = viper.BindEnv("port")
	_ = viper.BindPFlag("port", cmdFlags.Lookup("port"))

	_ = viper.BindEnv("api_key")
	// Storage config
	_ = viper.BindPFlag("storage", cmdFlags.Lookup("storage"))
	_ = viper.BindEnv("storage")
	rootCmd.AddCommand(cmdServe)
}

// @title Torrentd API
// @description Torrentd is a torrent RSS feed generator.
// @version 1.0.0
// @BasePath /
// @license MIT
// @host localhost:5000
// @schemes http
func serve(c *cobra.Command, _ []string) {
	facade, err := indexer.NewFacadeFromConfiguration(&appConfig)
	if err != nil {
		fmt.Printf("Couldn't initialize: %s", err)
		os.Exit(1)
	}
	// Init the server
	rserver := server.NewServer(&appConfig)
	err = rserver.Listen(facade)
	if err != nil {
		fmt.Print(err)
	}
}
