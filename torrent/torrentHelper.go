package torrent

import (
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/config"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/torznab"
)

type TorrentHelper struct {
	BasicTracker
	//pageSize uint
	indexer indexer.Indexer
}

func NewTorrentHelper(config config.Config) *TorrentHelper {
	rt := TorrentHelper{}
	ixr := config.GetString("indexer")
	if ixr == "" {
		ixr = "rutracker.org"
	}
	ixrObj, err := indexer.Lookup(config, ixr)
	if err != nil {
		log.Errorf("Could not find indexer `%s`.\n", ixr)
		return nil
	}
	rt.indexer = ixrObj
	return &rt
}

//func (r *TorrentHelper) clearSearch() {
//	//r.id = ""
//	//r.doc = nil
//	r.currentSearch = nil
//}

//Open the search to a given page.
func (th *TorrentHelper) Search(searchContext search.Instance, query string, page uint) (search.Instance, error) {
	qrobj := torznab.ParseQueryString(query)
	qrobj.Page = page
	srch, err := th.indexer.Search(qrobj, searchContext)
	if err != nil {
		return nil, err
	}
	return srch, nil
}

func (th *TorrentHelper) GetDefaultOptions() *GenericSearchOptions {
	return &GenericSearchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}

//Parse the torrent row
//func (th *TorrentHelper) parseTorrentRow(row *goquery.Selection) *search.ExternalResultItem {
//nameData := row.Find("a.tLink").Nodes[0].FirstChild.Data
//if nameData == "" {
//	return nil
//}
////Get the id of the rorrent
//torrentId, _ := row.Find("a.tLink").First().Attr("href")
//torrentId = formatting.ExtractAttr(torrentId, "t")
////Get the time on which the torrent was created
//torrentTime := formatting.FormatTime(formatting.ClearSpaces(row.Find("td").Last().Text()))
//
////Get the author
//authorNode := row.Find("td").Eq(4).Find("a").First()
//author := authorNode.Text()
//authorId, _ := authorNode.Attr("href")
//authorId = formatting.ExtractAttr(authorId, "pid")
////Get the category
//categoryNode := row.Find("td").Eq(2).Find("a").First()
//category := categoryNode.Text()
//categoryId, _ := categoryNode.Attr("href")
//categoryId = formatting.ExtractAttr(categoryId, "f")
////Get the size
//sizeNode := row.Find("td").Eq(5)
//size := formatting.SizeStrToBytes(sizeNode.Text())
////Get the downloads
//downloadsNode := row.Find("td").Eq(8)
//downloads, _ := strconv.Atoi(formatting.StripToNumber(downloadsNode.Text()))
////Get the leachers
//leachersTxt := formatting.StripToNumber(formatting.ClearSpaces(row.Find("td").Eq(7).Text()))
//leachers, _ := strconv.Atoi(leachersTxt)
////Get the seeders
//seedersNode := formatting.StripToNumber(formatting.ClearSpaces(row.Find("td").Eq(6).Text()))
//seeders, _ := strconv.Atoi(seedersNode)
//newTorrent := &search.ExternalResultItem{
//	ResultItem: search.ResultItem{
//		Title:       nameData,
//		PublishDate: torrentTime.Unix(),
//		Author:      author,
//		AuthorId:    authorId,
//		Size:        size,
//		Seeders:     seeders,
//		Peers:       leachers,
//		Grabs:       downloads,
//	},
//	LocalId:           torrentId,
//	LocalCategoryName: category,
//	LocalCategoryID:   categoryId,
//}
//newTorrent.IsMagnet = false
//if th.FetchDefinition {
//	def, err := ParseTorrentFromUrl(th, newTorrent.SourceLink)
//	if err != nil {
//		log.Warningf("Could not get torrent definition: %v", err)
//	} else {
//		newTorrent.Announce = def.Announce
//		newTorrent.Publisher = def.Publisher
//		newTorrent.Title = def.Info.Name
//		newTorrent.Size = def.GetTotalFileSize()
//	}
//}
//return newTorrent
//}
