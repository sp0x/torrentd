package torrent

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/config"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/formatting"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/torznab"
	"strconv"
)

type TorrentHelper struct {
	BasicTracker
	pageSize uint
	indexer  indexer.Indexer
}

func NewTorrentHelper(config config.Config) *TorrentHelper {
	rt := TorrentHelper{}
	key := "rutracker.org"
	ixr, err := indexer.Lookup(config, key)
	if err != nil {
		log.Errorf("Could not find indexer `%s`.\n", key)
		return nil
	}
	rt.indexer = ixr
	return &rt
}

//func (r *TorrentHelper) clearSearch() {
//	//r.id = ""
//	//r.doc = nil
//	r.currentSearch = nil
//}

//Open the search to a given page.
func (th *TorrentHelper) Search(searchContext *search.Search, query string, page uint) (*search.Search, error) {
	qrobj := torznab.ParseQueryString(query)
	qrobj.Page = page
	srch, err := th.indexer.Search(qrobj, searchContext)
	if err != nil {
		return nil, err
	}
	return srch, nil
	//
	////Get the page ids.
	//if searchContext == nil {
	//	searchContext, err = th.startSearch(query)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	//if page == 0 || searchContext == nil {
	//	return searchContext, nil // th.searchPage
	//}
	//furl := fmt.Sprintf("https://rutracker.org/forum/tracker.php?nm=%s&search_id=%s&start=%d", query, searchContext.Id, page*th.pageSize)
	//data, err := th.request(furl, nil, nil)
	//if err != nil {
	//	return nil, err
	//}
	//contentReader := bytes.NewReader(data)
	//doc, err := goquery.NewDocumentFromReader(contentReader)
	//if err != nil {
	//	return nil, err
	//}
	//searchContext.DOM = doc.First()
	////th.currentSearch.doc = doc
	//return searchContext, nil
}

//
////Start the search, getting the page ids
//func (th *TorrentHelper) startSearch(query string) (*search.Search, error) {
//	data := "prev_my=0&prev_new=0&prev_oop=0&o=1&s=2&tm=-1&pn=&submit=%CF%EE%E8%F1%EA"
//	for _, forumId := range []int{
//		//46, 56, 98, 103, 249, 314, 500, 552, 709, 1260, 2076, 2123, 2139,
//	} {
//		data += "&f%5B%5D=" + strconv.Itoa(forumId)
//	}
//	data += "&nm=" + query
//	furl := fmt.Sprintf("https://rutracker.org/forum/tracker.php?%s", data)
//	page, err := th.request(furl, nil, nil)
//	if err != nil {
//		return nil, err
//	}
//	contentReader := bytes.NewReader(page)
//	doc, err := goquery.NewDocumentFromReader(contentReader)
//	if err != nil {
//		return nil, err
//	}
//	/*
//		Scan all pages every time. It's not safe to skip them by last torrent ID in the database,
//		because some of them might be hidden at the previous run.
//	*/
//	pageUrlRx, _ := regexp.Compile("tracker.php\\?search_id=([^&]+)[^'\"]*?")
//	pageUrls := doc.Find("a.pg").FilterFunction(func(i int, s *goquery.Selection) bool {
//		//get href args that match tracker.php\\?search_id=([^&]+)[^'\"]*?
//		href, exists := s.Attr("href")
//		if !exists {
//			return false
//		}
//		matches := pageUrlRx.MatchString(href)
//		return matches
//	}).Map(func(i int, s *goquery.Selection) string {
//		href, _ := s.Attr("href")
//		matches := pageUrlRx.FindAllStringSubmatch(href, -1)
//		if len(matches) == 0 || len(matches[0]) < 2 {
//			return ""
//		}
//		return matches[0][1]
//	})
//	if len(pageUrls) == 0 {
//		lowerPage := strings.ToLower(string(page))
//		for _, reason := range []string{"форум временно отключен", "форум временно недоступен"} {
//			if strings.Contains(lowerPage, reason) {
//				return nil, errors.New("source in maintenance")
//			}
//		}
//		return nil, fmt.Errorf("no search pages found")
//	}
//	srch := search.Search{
//		DOM: doc.First(),
//		Id:  pageUrls[0],
//	}
//	//th.currentSearch = &search
//	return &srch, nil
//}

func (th *TorrentHelper) GetDefaultOptions() *GenericSearchOptions {
	return &GenericSearchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}

//Parse the torrent row
func (th *TorrentHelper) parseTorrentRow(row *goquery.Selection) *search.ExternalResultItem {
	nameData := row.Find("a.tLink").Nodes[0].FirstChild.Data
	if nameData == "" {
		return nil
	}
	//Get the id of the rorrent
	torrentId, _ := row.Find("a.tLink").First().Attr("href")
	torrentId = formatting.ExtractAttr(torrentId, "t")
	//Get the time on which the torrent was created
	torrentTime := formatting.FormatTime(formatting.ClearSpaces(row.Find("td").Last().Text()))

	//Get the author
	authorNode := row.Find("td").Eq(4).Find("a").First()
	author := authorNode.Text()
	authorId, _ := authorNode.Attr("href")
	authorId = formatting.ExtractAttr(authorId, "pid")
	//Get the category
	categoryNode := row.Find("td").Eq(2).Find("a").First()
	category := categoryNode.Text()
	categoryId, _ := categoryNode.Attr("href")
	categoryId = formatting.ExtractAttr(categoryId, "f")
	//Get the size
	sizeNode := row.Find("td").Eq(5)
	size := formatting.SizeStrToBytes(sizeNode.Text())
	//Get the downloads
	downloadsNode := row.Find("td").Eq(8)
	downloads, _ := strconv.Atoi(formatting.StripToNumber(downloadsNode.Text()))
	//Get the leachers
	leachersTxt := formatting.StripToNumber(formatting.ClearSpaces(row.Find("td").Eq(7).Text()))
	leachers, _ := strconv.Atoi(leachersTxt)
	//Get the seeders
	seedersNode := formatting.StripToNumber(formatting.ClearSpaces(row.Find("td").Eq(6).Text()))
	seeders, _ := strconv.Atoi(seedersNode)
	newTorrent := &search.ExternalResultItem{
		ResultItem: search.ResultItem{
			Title:       nameData,
			PublishDate: torrentTime.Unix(),
			Author:      author,
			AuthorId:    authorId,
			Size:        size,
			Seeders:     seeders,
			Peers:       leachers,
			Grabs:       downloads,
		},
		LocalId:           torrentId,
		LocalCategoryName: category,
		LocalCategoryID:   categoryId,
	}
	newTorrent.Link = th.GetTorrentLink(newTorrent)
	newTorrent.SourceLink = th.GetTorrentDownloadLink(newTorrent)
	newTorrent.IsMagnet = false
	if th.FetchDefinition {
		def, err := ParseTorrentFromUrl(th, newTorrent.SourceLink)
		if err != nil {
			log.Warningf("Could not get torrent definition: %v", err)
		} else {
			newTorrent.Announce = def.Announce
			newTorrent.Publisher = def.Publisher
			newTorrent.Title = def.Info.Name
			newTorrent.Size = def.GetTotalFileSize()
		}
	}
	return newTorrent
}

func (th *TorrentHelper) GetTorrentLink(t *search.ExternalResultItem) string {
	return fmt.Sprintf("http://rutracker.org/forum/viewtopic.php?t=%s", t.LocalId)
}

func (th *TorrentHelper) GetTorrentDownloadLink(t *search.ExternalResultItem) string {
	return fmt.Sprintf("http://rutracker.org/forum/dl.php?t=%s", t.LocalId)
}

//func (th *TorrentHelper) ParseTorrents(doc *goquery.Selection, f func(i int, s *search.ExternalResultItem)) *goquery.Selection {
//	return doc.Find("tr.tCenter.hl-tr").Each(func(i int, s *goquery.Selection) {
//		torrent := th.parseTorrentRow(s)
//		f(i, torrent)
//	})
//}
