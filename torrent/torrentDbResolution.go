package torrent

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/db"
)

func ResolveTorrents(client *Rutracker, hours int) {
	gdb := db.GetOrmDb()
	defer gdb.Close()
	torrents := client.storage.GetOlderThanHours(hours)
	for _, t := range torrents {
		def, err := ParseTorrentFromUrl(client, t.DownloadLink)
		if err != nil {
			log.Debugf("Could not resolve torrent: [%v] %v", t.TorrentId, t.Name)
			continue
		}
		t.Announce = def.Announce
		t.Publisher = def.Publisher
		t.Name = def.Info.Name
		t.Size = def.GetTotalFileSize()
		gdb.Save(t)
	}
}
