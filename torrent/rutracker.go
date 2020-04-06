package torrent

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/requests"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Rutracker struct {
	client              *http.Client
	loggedIn            bool
	lastRequest         time.Time
	currentSearchPageId string
	currentSearchDoc    *goquery.Document
	pageSize            uint
	torrentStorage      *Storage
}

func NewRutracker() *Rutracker {
	rt := Rutracker{}
	rt.pageSize = 50
	rt.torrentStorage = &Storage{}
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		DisableCompression: false,
	}
	rt.client = &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
		Jar:       jar,
	}
	return &rt
}

//Login with the tracker.
func (r *Rutracker) Login(username, password string) error {
	loginUrl := "https://rutracker.org/forum/login.php"
	if username == "" || password == "" {
		return errors.New("no auth credentials given")
	}
	data := []byte(fmt.Sprintf("login_username=%s&login_password=%s&login=%C2%F5%EE%E4", username, password))
	page, err := requests.Post(r.client, loginUrl, data, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		return err
	}
	page = DecodeWindows1251(page)
	r.loggedIn = true
	return nil
}

//Send a request to the tracker
func (r *Rutracker) request(urlx string, data []byte, headers map[string]string) ([]byte, error) {
	maxPerSecond := 1
	minDiff := 1.0 / maxPerSecond
	timeElapsed := time.Now().Sub(r.lastRequest)
	if int(timeElapsed.Seconds()) < int(minDiff) {
		t := r.lastRequest.Add(time.Second * time.Duration(minDiff)).Sub(time.Now())
		time.Sleep(t)
	}
	r.lastRequest = time.Now()
	resp, err := requests.Post(r.client, urlx, data, headers)
	if err != nil {
		return nil, err
	}
	resp = DecodeWindows1251(resp)
	return resp, nil
}

//Open the search to a given page.
func (r *Rutracker) search(page uint) (*goquery.Document, error) {
	if !r.loggedIn {
		return nil, errors.New("not logged in")
	}
	var searchDoc *goquery.Document
	var err error
	//Get the page ids.
	if r.currentSearchPageId == "" {
		searchDoc, err = r.startSearch()
		if err != nil {
			return nil, err
		}
	}
	if page == 0 {
		return searchDoc, nil // r.searchPage
	}
	furl := fmt.Sprintf("https://rutracker.org/forum/tracker.php?search_id=%s&start=%d", r.currentSearchPageId, page*r.pageSize)
	data, err := r.request(furl, nil, nil)
	if err != nil {
		return nil, err
	}
	contentReader := bytes.NewReader(data)
	doc, err := goquery.NewDocumentFromReader(contentReader)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

//Start the search, getting the page ids
func (r *Rutracker) startSearch() (*goquery.Document, error) {
	data := "prev_my=0&prev_new=0&prev_oop=0&o=1&s=2&tm=-1&pn=&nm=&submit=%CF%EE%E8%F1%EA"
	for _, forumId := range []int{
		46, 56, 98, 103, 249, 314, 500, 552, 709, 1260, 2076, 2123, 2139,
	} {
		data += "&f%5B%5D=" + strconv.Itoa(forumId)
	}
	page, err := r.request("https://rutracker.org/forum/tracker.php", []byte(data), nil)
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
	r.currentSearchPageId = pageUrls[0]
	r.currentSearchDoc = doc
	return doc, nil
}

func (r *Rutracker) getDefaultOptions() *FetchOptions {
	return &FetchOptions{
		PageCount:            10,
		StartingPage:         0,
		MaxRequestsPerSecond: 1,
	}
}

func (r *Rutracker) parseTorrentRow(row *goquery.Selection) *db.Torrent {
	nameData := row.Find("a.tLink").Nodes[0].FirstChild.Data
	if nameData == "" {
		return nil
	}
	//Get the id of the rorrent
	torrentId, _ := row.Find("a.tLink").First().Attr("href")
	torrentId = extractAttr(torrentId, "t")
	//Get the time on which the torrent was created
	torrentTime := clearSpaces(row.Find("td").Last().Text())
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
		AddedOn:      torrentTime,
		AuthorName:   author,
		AuthorId:     authorId,
		CategoryName: category,
		CategoryId:   categoryId,
		Size:         size,
		Seeders:      seeders,
		Leachers:     leachers,
		Downloaded:   downloads,
	}
	newTorrent.Link = r.getTorrentLink(newTorrent)
	return newTorrent
}

func (r *Rutracker) getTorrentLink(t *db.Torrent) string {
	return fmt.Sprintf("http://rutracker.org/forum/viewtopic.php?t=%s", t.TorrentId)
}

func (r *Rutracker) parseTorrents(doc *goquery.Document, f func(i int, s *db.Torrent)) *goquery.Selection {
	return doc.Find("tr.tCenter.hl-tr").Each(func(i int, s *goquery.Selection) {
		torrent := r.parseTorrentRow(s)
		f(i, torrent)
	})
}
