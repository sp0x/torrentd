package torrent

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"reflect"
)

//Gets torrent information from a given tracker and updates the torrent db
func ResolveTorrents(index indexer.Indexer, config config.Config) []search.ResultItemBase {
	store := storage.NewBuilder().
		WithRecord(&search.ScrapeResultItem{}).
		Build()
	defer store.Close()
	results := store.GetLatest(20)
	if err := index.Check(); err != nil {
		log.Errorf("Failed while checking indexer %s. Err: %s\n", reflect.TypeOf(index), err)
		return nil
	}
	indexScope := indexer.NewScope()
	for i, searchItem := range results {
		//Skip already resolved results.
		item := searchItem.(*search.TorrentResultItem)
		if item.Announce != "" {
			continue
		}
		index, err := indexScope.Lookup(config, item.Site)
		if err != nil {
			log.WithFields(log.Fields{"err": err, "site": item.Site}).
				Warningf("Error while looking up indexer.")
			continue
		}
		if index == nil {
			log.WithFields(log.Fields{"site": item.Site}).
				Warningf("Couldn'item find indexer.")
			continue
		}
		err = index.Check()
		if err != nil {
			log.WithFields(log.Fields{"err": err, "site": item.Site}).
				Warningf("Error while checking indexer.")
			continue
		}
		responsePxy, err := index.Open(item)
		if err != nil {
			log.Debugf("Couldn'item open result [%v] %v", item.LocalId, item.Title)
			continue
		}
		log.
			WithFields(log.Fields{"link": item.SourceLink, "name": item.Title}).
			Info("Resolving")
		def, err := ParseTorrentFromStream(responsePxy.Reader)
		if err != nil {
			log.Debugf("Could not resolve result: [%v] %v", item.LocalId, item.Title)
			continue
		}
		item.Announce = def.Announce
		item.Publisher = def.Publisher
		item.OriginalTitle = def.Info.Name
		item.Size = def.GetTotalFileSize()
		item.PublishedWith = def.CreatedBy
		perc := (float32(i) / float32(len(results))) * 100
		log.WithFields(log.Fields{"id": item.LocalId, "title": item.Title}).
			Infof("%f%% Resolved ", perc)
		err = store.Add(item)
		if err != nil {
			log.Errorf("Could not save result: %v", err)
		}
	}
	return results
}
