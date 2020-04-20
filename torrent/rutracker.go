package torrent

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/indexer"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/requests"
	"regexp"
	"strconv"
	"strings"
)

type Rutracker struct {
	BasicTracker
	loggedIn bool
	//id string
	//doc    *goquery.DOM
	//currentSearch *Search
	pageSize uint
}

func NewRutracker() *Rutracker {
	rt := Rutracker{}
	rt.pageSize = 50
	rt.client = newWebClient()
	return &rt
}

//Login with the tracker.
func (r *Rutracker) Login(username, password string) error {
	loginUrl := "https://rutracker.org/forum/login.php"
	if username == "" || password == "" {
		return errors.New("no auth credentials given")
	}
	data := []byte(fmt.Sprintf("login_username=%s&login_password=%s%s", username, password, "&login=%C2%F5%EE%E4"))
	log.Debugf("Trying to login to: %s with user %s", loginUrl, username)
	page, err := requests.Post(r.client, loginUrl, data, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		return err
	}
	page = indexer.DecodeWindows1251(page)
	r.loggedIn = true
	return nil
}

//func (r *Rutracker) clearSearch() {
//	//r.id = ""
//	//r.doc = nil
//	r.currentSearch = nil
//}

//Open the search to a given page.
func (r *Rutracker) Search(searchContext *search.Search, query string, page uint) (*search.Search, error) {
	if !r.loggedIn {
		return nil, errors.New("not logged in")
	}

	var err error
	//Get the page ids.
	if searchContext == nil {
		searchContext, err = r.startSearch(query)
		if err != nil {
			return nil, err
		}
	}
	if page == 0 || searchContext == nil {
		return searchContext, nil // r.searchPage
	}
	furl := fmt.Sprintf("https://rutracker.org/forum/tracker.php?nm=%s&search_id=%s&start=%d", query, searchContext.Id, page*r.pageSize)
	data, err := r.request(furl, nil, nil)
	if err != nil {
		return nil, err
	}
	contentReader := bytes.NewReader(data)
	doc, err := goquery.NewDocumentFromReader(contentReader)
	if err != nil {
		return nil, err
	}
	searchContext.DOM = doc.First()
	//r.currentSearch.doc = doc
	return searchContext, nil
}

//Start the search, getting the page ids
func (r *Rutracker) startSearch(query string) (*search.Search, error) {
	data := "prev_my=0&prev_new=0&prev_oop=0&o=1&s=2&tm=-1&pn=&submit=%CF%EE%E8%F1%EA"
	for _, forumId := range []int{
		//46, 56, 98, 103, 249, 314, 500, 552, 709, 1260, 2076, 2123, 2139,
	} {
		data += "&f%5B%5D=" + strconv.Itoa(forumId)
	}
	data += "&nm=" + query
	furl := fmt.Sprintf("https://rutracker.org/forum/tracker.php?%s", data)
	page, err := r.request(furl, nil, nil)
	if err != nil {
		return nil, err
	}
	contentReader := bytes.NewReader(page)
	doc, err := goquery.NewDocumentFromReader(contentReader)
	if err != nil {
		return nil, err
	}
	/*
		Scan all pages every time. It's not safe to skip them by last torrent ID in the database,
		because some of them might be hidden at the previous run.
	*/
	pageUrlRx, _ := regexp.Compile("tracker.php\\?search_id=([^&]+)[^'\"]*?")
	pageUrls := doc.Find("a.pg").FilterFunction(func(i int, s *goquery.Selection) bool {
		//get href args that match tracker.php\\?search_id=([^&]+)[^'\"]*?
		href, exists := s.Attr("href")
		if !exists {
			return false
		}
		matches := pageUrlRx.MatchString(href)
		return matches
	}).Map(func(i int, s *goquery.Selection) string {
		href, _ := s.Attr("href")
		matches := pageUrlRx.FindAllStringSubmatch(href, -1)
		if len(matches) == 0 || len(matches[0]) < 2 {
			return ""
		}
		return matches[0][1]
	})
	if len(pageUrls) == 0 {
		lowerPage := strings.ToLower(string(page))
		for _, reason := range []string{"форум временно отключен", "форум временно недоступен"} {
			if strings.Contains(lowerPage, reason) {
				return nil, errors.New("source in maintenance")
			}
		}
		return nil, fmt.Errorf("no search pages found")
	}
	srch := search.Search{
		DOM: doc.First(),
		Id:  pageUrls[0],
	}
	//r.currentSearch = &search
	return &srch, nil
}

func (r *Rutracker) GetDefaultOptions() *FetchOptions {
	return &FetchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}

//Parse the torrent row
func (r *Rutracker) parseTorrentRow(row *goquery.Selection) *db.Torrent {
	nameData := row.Find("a.tLink").Nodes[0].FirstChild.Data
	if nameData == "" {
		return nil
	}
	//Get the id of the rorrent
	torrentId, _ := row.Find("a.tLink").First().Attr("href")
	torrentId = extractAttr(torrentId, "t")
	//Get the time on which the torrent was created
	torrentTime := formatTime(clearSpaces(row.Find("td").Last().Text()))

	//Get the author
	authorNode := row.Find("td").Eq(4).Find("a").First()
	author := authorNode.Text()
	authorId, _ := authorNode.Attr("href")
	authorId = extractAttr(authorId, "pid")
	//Get the category
	categoryNode := row.Find("td").Eq(2).Find("a").First()
	category := categoryNode.Text()
	categoryId, _ := categoryNode.Attr("href")
	categoryId = extractAttr(categoryId, "f")
	//Get the size
	sizeNode := row.Find("td").Eq(5)
	size := sizeStrToBytes(sizeNode.Text())
	//Get the downloads
	downloadsNode := row.Find("td").Eq(8)
	downloads, _ := strconv.Atoi(stripToNumber(downloadsNode.Text()))
	//Get the leachers
	leachersTxt := stripToNumber(clearSpaces(row.Find("td").Eq(7).Text()))
	leachers, _ := strconv.Atoi(leachersTxt)
	//Get the seeders
	seedersNode := stripToNumber(clearSpaces(row.Find("td").Eq(6).Text()))
	seeders, _ := strconv.Atoi(seedersNode)
	newTorrent := &db.Torrent{
		Name:         nameData,
		TorrentId:    torrentId,
		AddedOn:      torrentTime.Unix(),
		AuthorName:   author,
		AuthorId:     authorId,
		CategoryName: category,
		CategoryId:   categoryId,
		Size:         size,
		Seeders:      seeders,
		Leachers:     leachers,
		Downloaded:   downloads,
	}
	newTorrent.Link = r.GetTorrentLink(newTorrent)
	newTorrent.DownloadLink = r.GetTorrentDownloadLink(newTorrent)
	newTorrent.IsMagnet = false
	if r.FetchDefinition {
		def, err := ParseTorrentFromUrl(r, newTorrent.DownloadLink)
		if err != nil {
			log.Warningf("Could not get torrent definition: %v", err)
		} else {
			newTorrent.Announce = def.Announce
			newTorrent.Publisher = def.Publisher
			newTorrent.Name = def.Info.Name
			newTorrent.Size = def.GetTotalFileSize()
		}
	}
	return newTorrent
}

func (r *Rutracker) GetTorrentLink(t *db.Torrent) string {
	return fmt.Sprintf("http://rutracker.org/forum/viewtopic.php?t=%s", t.TorrentId)
}

func (r *Rutracker) GetTorrentDownloadLink(t *db.Torrent) string {
	return fmt.Sprintf("http://rutracker.org/forum/dl.php?t=%s", t.TorrentId)
}

func (r *Rutracker) ParseTorrents(doc *goquery.Selection, f func(i int, s *db.Torrent)) *goquery.Selection {
	return doc.Find("tr.tCenter.hl-tr").Each(func(i int, s *goquery.Selection) {
		torrent := r.parseTorrentRow(s)
		f(i, torrent)
	})
}
