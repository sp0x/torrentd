package torrent

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/storage/sqlite"
	"os"
	"reflect"
	"text/tabwriter"
)

//Gets torrent information from a given tracker and updates the torrent db
func ResolveTorrents(client *indexer.Facade, hours int) {
	dbStorage := sqlite.DBStorage{}
	torrents := dbStorage.GetOlderThanHours(hours)
	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)
	if err := client.Indexer.Check(); err != nil {
		log.Errorf("Failed while checking indexer %s. Err: %s\n", reflect.TypeOf(client.Indexer), err)
		return
	}
	scope := indexer.NewScope()
	for i, t := range torrents {
		//Skip already resolved torrents.
		if t.Announce != "" {
			continue
		}
		ixr, err := scope.Lookup(client.Config, t.Site)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "site": t.Site}).
				Warningf("Error while looking up indexer.")
			continue
		}
		if ixr == nil {
			log.WithFields(log.Fields{"site": t.Site}).
				Warningf("Couldn't find indexer.")
			continue
		}
		err = ixr.Check()
		if err != nil {
			log.WithFields(log.Fields{"err": err, "site": t.Site}).
				Warningf("Error while checking indexer.")
			continue
		}
		reader, err := ixr.Open(&t)
		if err != nil {
			log.Debugf("Couldn't open result [%v] %v", t.LocalId, t.Title)
			continue
		}
		log.
			WithFields(log.Fields{"link": t.SourceLink, "name": t.Title}).
			Info("Resolving")
		def, err := ParseTorrentFromStream(reader)
		if err != nil {
			log.Debugf("Could not resolve result: [%v] %v", t.LocalId, t.Title)
			continue
		}
		t.Announce = def.Announce
		t.Publisher = def.Publisher
		t.OriginalTitle = def.Info.Name
		t.Size = def.GetTotalFileSize()
		t.PublishedWith = def.CreatedBy
		perc := (float32(i) / float32(len(torrents))) * 100
		_, _ = fmt.Fprintf(tabWr, "%f%% Resolved [%s]\t%s\n", perc, t.LocalId, t.Title)
		err = dbStorage.Create(nil, &t)
		if err != nil {
			log.Errorf("Could not save result: %v", err)
		}
		_ = tabWr.Flush()
	}
}
