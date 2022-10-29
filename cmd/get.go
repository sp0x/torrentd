package main

import (
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/status"
	"github.com/sp0x/torrentd/storage/bolt"
)

func init() {
	cmdGet := &cobra.Command{
		Use:   "get",
		Short: "Run a query or get results from index(es)",
		Run:   getCommand,
	}
	storage := ""
	storageEndpoint := ""
	query := ""
	workers := 0
	users := 1
	cmdFlags := cmdGet.PersistentFlags()
	cmdFlags.StringVarP(&storage, "storage", "o", "boltdb", `The storage backing to use.
Currently supported storage backings: boltdb, firebase, sqlite`)
	cmdFlags.StringVarP(&storageEndpoint, "storageendpoint", "s", bolt.GetDefaultDatabasePath(), `The endpoint that should be used for storing data.`)
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
	_ = viper.BindEnv("storageendpoint")
	_ = viper.BindPFlag("storageendpoint", cmdFlags.Lookup("storageendpoint"))
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
	status.SetupPubsub(appConfig.GetString("firebase_project"))
	facade := indexer.NewFacadeFromConfiguration(&appConfig)
	if facade == nil {
		log.Error("Couldn't initialize index facade.")
		return
	}

	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	queryStr := c.Flag("query").Value.String()
	query, _ := search.NewQueryFromQueryString(queryStr)
	query.StopOnStale = true
	if query.NumberOfPagesToFetch == 0 {
		query.NumberOfPagesToFetch = 20
	}
	results, err := facade.Search(query)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	for resultsBatch := range results {
		indexer.PrintResults(resultsBatch, tabWr)
		_ = tabWr.Flush()
	}
}
