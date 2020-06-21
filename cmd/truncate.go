package main

import (
	"github.com/sp0x/torrentd/storage/sqlite"
	"github.com/spf13/cobra"
)

func init() {
	cmdTruncate := &cobra.Command{
		Use:     "truncate",
		Aliases: []string{"t"},
		Short:   "Truncates the database",
		Run:     truncateTorrentDb,
	}
	rootCmd.AddCommand(cmdTruncate)
}

func truncateTorrentDb(cmd *cobra.Command, args []string) {
	storage := sqlite.DBStorage{}
	storage.Truncate()
}
