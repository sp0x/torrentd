package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/status"
)

func init() {
	cmdGet := &cobra.Command{
		Use:   "get",
		Short: "Run a query or get results from index(es)",
		Run:   getCommand,
	}
	storage := ""
	query := ""
	workers := 0
	users := 1
	cmdFlags := cmdGet.PersistentFlags()
	cmdFlags.StringVarP(&storage, "storage", "o", "boltdb", `The storage backing to use.
Currently supported storage backings: boltdb, firebase, sqlite`)
	cmdFlags.StringVar(&query, "query", "", `Query to use when searching`)
	cmdFlags.IntVar(&workers, "workers", 0, "The number of parallel searches that can be used.")
	cmdFlags.IntVar(&users, "users", 1, "The number of user sessions to use in rotation.")
	_ = viper.BindEnv("workers")
	_ = viper.BindEnv("users")
	firebaseProject := ""
	firebaseCredentials := ""
	cmdFlags.StringVarP(&firebaseCredentials, "firebase_project", "", "", "The project id for firebase")
	cmdFlags.StringVarP(&firebaseProject, "firebase_credentials_file", "", "", "The service credentials for firebase")
	// Storage config
	_ = viper.BindPFlag("storage", cmdFlags.Lookup("storage"))
	_ = viper.BindEnv("storage")
	// Firebase related
	_ = viper.BindPFlag("firebase_project", cmdFlags.Lookup("firebase_project"))
	_ = viper.BindEnv("firebase_project")
	_ = viper.BindPFlag("firebase_credentials_file", cmdFlags.Lookup("firebase_credentials_file"))
	_ = viper.BindEnv("firebase_credentials_file")

	_ = viper.BindPFlags(cmdFlags)
	_ = viper.BindEnv("query")
	rootCmd.AddCommand(cmdGet)
}

func getCommand(c *cobra.Command, _ []string) {
	facade := indexer.NewFacadeFromConfiguration(&appConfig)
	if facade == nil {
		log.Error("Couldn't initialize torrent facade.")
		return
	}
	// Start watching the torrent tracker.
	status.SetupPubsub(appConfig.GetString("firebase_project"))
	queryStr := c.Flag("query").Value.String()
	query, _ := search.NewQueryFromQueryString(queryStr)
	err := indexer.Get(facade, query)
	if err != nil {
		fmt.Printf("Couldn't get results: ")
		os.Exit(1)
	}
}
