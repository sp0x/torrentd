package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/status"
	"github.com/sp0x/torrentd/torznab"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

func init() {
	cmdGet := &cobra.Command{
		Use:   "get",
		Short: "Run a query or get results from index(es)",
		Run:   getCommand,
	}
	storage := ""
	query := ""
	cmdFlags := cmdGet.Flags()
	cmdFlags.StringVarP(&storage, "storage", "o", "boltdb", `The storage backing to use.
Currently supported storage backings: boltdb, firebase, sqlite`)
	cmdFlags.StringVarP(&query, "query", "", "", `Query to use when searching`)
	firebaseProject := ""
	firebaseCredentials := ""
	cmdFlags.StringVarP(&firebaseCredentials, "firebase_project", "", "", "The project id for firebase")
	cmdFlags.StringVarP(&firebaseProject, "firebase_credentials_file", "", "", "The service credentials for firebase")
	viper.SetDefault("port", 5000)
	_ = viper.BindEnv("port")
	_ = viper.BindEnv("api_key")
	//Storage config
	_ = viper.BindPFlag("storage", cmdFlags.Lookup("storage"))
	_ = viper.BindEnv("storage")
	//Firebase related
	_ = viper.BindPFlag("firebase_project", cmdFlags.Lookup("firebase_project"))
	_ = viper.BindEnv("firebase_project")
	_ = viper.BindPFlag("firebase_credentials_file", cmdFlags.Lookup("firebase_credentials_file"))
	_ = viper.BindEnv("firebase_credentials_file")
	_ = viper.BindPFlag("query", cmdFlags.Lookup("query"))
	_ = viper.BindEnv("query")
	rootCmd.AddCommand(cmdGet)
}

func getCommand(_ *cobra.Command, _ []string) {
	facade := indexer.NewFacadeFromConfiguration(&appConfig)
	if facade == nil {
		log.Error("Couldn't initialize torrent facade.")
		return
	}
	//Start watching the torrent tracker.
	status.SetupPubsub(appConfig.GetString("firebase_project"))
	query := torznab.ParseQueryString(viper.GetString("query"))
	err := indexer.Get(facade, query)
	if err != nil {
		fmt.Printf("Couldn't get results: ")
		os.Exit(1)
	}
}
