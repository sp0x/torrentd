package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/status"
)

func init() {
	cmdWatch := &cobra.Command{
		Use:   "watch",
		Short: "Watches the torrent tracker for new torrents.",
		Run:   watchIndex,
	}
	storage := ""
	query := ""
	cmdFlags := cmdWatch.Flags()
	watchInterval := 0
	cmdFlags.IntVarP(&watchInterval, "interval", "i", 10, "Interval between checks.")
	cmdFlags.StringVarP(&storage, "storage", "o", "boltdb", `The storage backing to use.
Currently supported storage backings: boltdb, firebase, sqlite`)
	cmdFlags.StringVarP(&query, "query", "", "", `Query to use when searching`)
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
	_ = viper.BindPFlag("query", cmdFlags.Lookup("query"))
	_ = viper.BindEnv("query")
	_ = viper.BindPFlag("interval", cmdFlags.Lookup("interval"))
	rootCmd.AddCommand(cmdWatch)
}

func watchIndex(c *cobra.Command, _ []string) {
	facade, err := indexer.NewFacadeFromConfiguration(&appConfig)
	if err != nil {
		fmt.Printf("Couldn't initialize: %s", err)
		return
	}

	// Start watching the torrent tracker.
	status.SetupPubsub(appConfig.GetString("firebase_project"))
	query, err := search.NewQueryFromQueryString(c.Flag("query").Value.String())
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	watchInterval := viper.GetInt("interval")
	resultChannel := indexer.Watch(facade, query, watchInterval)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)
	for item := range resultChannel {
		if !item.IsNew() && !item.IsUpdate() {
			continue
		}
		if item.IsNew() && !item.IsUpdate() {
			_, _ = fmt.Fprintf(tabWr, "Found new result #%s:\t%s\n",
				item.UUID(), item.String())
		} else {
			_, _ = fmt.Fprintf(tabWr, "Updated torrent #%s:\t%s\n",
				item.UUID(), item.String())
		}
	}
}
