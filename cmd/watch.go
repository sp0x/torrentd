package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"text/tabwriter"
)

var watchInterval int

func init() {
	cmdWatch := &cobra.Command{
		Use:   "watch",
		Short: "Watches the torrent tracker for new torrents.",
		Run:   watchTracker,
	}
	storage := ""
	cmdFlags := cmdWatch.Flags()
	cmdFlags.IntVarP(&watchInterval, "interval", "i", 10, "Interval between checks.")
	cmdFlags.StringVarP(&storage, "storage", "o", "boltdb", `The storage backing to use.
Currently supported storage backings: boltdb, firestore, sqlite`)
	viper.SetDefault("port", 5000)
	_ = viper.BindEnv("port")
	_ = viper.BindEnv("api_key")
	_ = viper.BindPFlag("storage", cmdFlags.Lookup("storage"))
	_ = viper.BindEnv("storage")
	rootCmd.AddCommand(cmdWatch)
}

func watchTracker(_ *cobra.Command, _ []string) {
	facade := indexer.NewFacadeFromConfiguration(&appConfig)
	if facade == nil {
		log.Error("Couldn't initialize torrent facade.")
		return
	}
	//Init the server
	go func() {
		rserver := server.NewServer(&appConfig)
		err := rserver.Listen(facade)
		if err != nil {
			fmt.Print(err)
		}
	}()

	//Start watching the torrent tracker.
	resultChannel := indexer.Watch(facade, nil, watchInterval)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)
	for item := range resultChannel {
		if !item.IsNew() && !item.IsUpdate() {
			continue
		}
		if item.IsNew() && !item.IsUpdate() {
			_, _ = fmt.Fprintf(tabWr, "Found new result #%s:\t%s\t[%s]:\t%s\n",
				item.LocalId, item.AddedOnStr(), item.Fingerprint, item.Title)
		} else {
			_, _ = fmt.Fprintf(tabWr, "Updated torrent #%s:\t%s\t[%s]:\t%s\n",
				item.LocalId, item.AddedOnStr(), item.Fingerprint, item.Title)
		}
	}
}
