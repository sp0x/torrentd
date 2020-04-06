package torrent

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/requests"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type Rutracker struct {
	client              *http.Client
	loggedIn            bool
	lastRequest         time.Time
	currentSearchPageId string
	currentSearchDoc    *goquery.Document
	pageSize            int
	torrentStorage      *TorrentStorage
}

func newRuTracker() *Rutracker {
	rt := Rutracker{}
	rt.pageSize = 50
	rt.torrentStorage = &TorrentStorage{}
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		DisableCompression: false,
	}
	rt.client = &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
		Jar:       jar, //Commented because this causes CSRF issues if enabled
	}
	return &rt
}

func (r *Rutracker) login(username, password string) error {
	loginUrl := "https://rutracker.org/forum/login.php"
	data := []byte(fmt.Sprintf("login_username=%s&login_password=%s&login=%C2%F5%EE%E4", username, password))
	//data2 := make([]byte, len(data)*2, len(data)*2)
	//ascii85.Encode(data2, data)
	page, err := requests.Post(r.client, loginUrl, data, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if err != nil {
		return err
	}
	page = DecodeWindows1251(page)
	r.loggedIn = true
	// log.Print(string(page))
	return nil
}

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
func (r *Rutracker) search(page int) (*goquery.Document, error) {
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

func GetNewTorrents(user, pass string) error {
	log.Info("Searching for new torrents")

	//tr1 := tagRegex("a", "tLink", "href", `viewtopic\.php\?t=(\d+)`)
	//tr2 := tagRegex("td", "row4", "data-ts_text", "(\\d+)")
	//torrentRegex, _ := regexp.Compile(tr1 + "(.+?)<a/>" +
	//	`.+` +
	//	tr2 +
	//	"\\s*<p>\\s*\\d{1,2}-(?:Янв|Фев|Мар|Апр|Май|Июн|Июл|Авг|Сен|Окт|Ноя|Дек)-\\d{2}</p>")
	client := newRuTracker()
	err := client.login(user, pass)
	if err != nil {
		log.Error("Could not log in.")
		return err
	}
	pagesToScan := 10
	torrentsPerPage := 50
	totalTorrents := pagesToScan * torrentsPerPage
	page := 0
	log.Print(totalTorrents)

	tabWr := new(tabwriter.Writer)
	tabWr.Init(os.Stdout, 0, 8, 0, '\t', 0)

	for page = 0; page < pagesToScan; page++ {
		log.Infof("Getting page %d\n", page)
		pageDoc, err := client.search(page)
		if err != nil {
			log.Warning("Could not fetch page %d", page)
			continue
		}
		/*
			Scan all pages every time. It's not safe to skip them by last torrent ID in the database,
			because some of them might be hidden at the previous run.
		*/
		counter := 0
		pageTorrentId := 0
		torrentIdRx, _ := regexp.Compile("\\?t=(\\d+)")
		pageDoc.Find("tr.tCenter.hl-tr").Each(func(i int, s *goquery.Selection) {
			torrentNumber := page*client.pageSize + pageTorrentId + 1
			nameData := s.Find("a.tLink").Nodes[0].FirstChild.Data
			if nameData == "" {
				return
			}
			//Get the id of the rorrent
			torrentId, _ := s.Find("a.tLink").First().Attr("href")
			idMatches := torrentIdRx.FindAllStringSubmatch(torrentId, -1)
			torrentId = idMatches[0][1]
			//Get the time on which the torrent was created
			torrentTime := clearSpaces(s.Find("td").Last().Text())
			//Get the author
			authorNode := s.Find("td").Eq(4).Find("a").First()
			author := authorNode.Text()
			authorId, _ := authorNode.Attr("href")
			authorId = extractAttr(authorId, "pid")
			//Get the category
			categoryNode := s.Find("td").Eq(2).Find("a").First()
			category := categoryNode.Text()
			categoryId, _ := categoryNode.Attr("href")
			categoryId = extractAttr(categoryId, "f")
			//Get the size
			sizeNode := s.Find("td").Eq(5)
			size := sizeStrToBytes(sizeNode.Text())
			//Get the downloads
			downloadsNode := s.Find("td").Eq(8)
			downloads, _ := strconv.Atoi(stripToNumber(downloadsNode.Text()))
			//Get the leachers
			leachersTxt := stripToNumber(clearSpaces(s.Find("td").Eq(7).Text()))
			leachers, _ := strconv.Atoi(leachersTxt)
			//Get the seeders
			seedersNode := stripToNumber(clearSpaces(s.Find("td").Eq(6).Text()))
			seeders, _ := strconv.Atoi(seedersNode)

			existingTorrent := client.torrentStorage.FindByTorrentId(torrentId)
			isNew := existingTorrent == nil || existingTorrent.AddedOn != torrentTime
			if isNew && torrentNumber >= totalTorrents/2 {
				log.Warningf("Got a new torrent after a half of the search (%d of %d). "+
					"Consider to increase the search page number.\n", torrentNumber, totalTorrents)
			}
			if isNew || (existingTorrent != nil && existingTorrent.Name != nameData) {
				newTorrent := db.Torrent{
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
				newTorrent.Fingerprint = getTorrentFingerprint(&newTorrent)
				newTorrent.Link = getTorrentLink(&newTorrent)
				_, _ = fmt.Fprintf(tabWr, "Found new torrent #%s:\t%s\t[%s]:\t%s\n",
					torrentId, newTorrent.AddedOn, newTorrent.Fingerprint, newTorrent.Name)
				client.torrentStorage.Create(&newTorrent)
			} else {
				_, _ = fmt.Fprintf(tabWr, "Torrent #%s:\t%s\t[%s]:\t%s\n",
					torrentId, torrentTime, "#", nameData)

			}
			_ = tabWr.Flush()
			counter++
		})
		if counter != client.pageSize {
			log.Errorf("Error while parsing page %s: got %s torrents instead of %s\n", page, counter, client.torrentStorage)
		}
		//matches := torrentRegex.FindAllString(pageContent, -1)
		//log.Print(matches)
	}
	return nil
}

func getTorrentLink(t *db.Torrent) string {
	return fmt.Sprintf("http://rutracker.org/forum/viewtopic.php?t=%s", t.TorrentId)
}

func tagRegex(tag, tagClass string, matchParam string, matchValue string) string {
	if matchValue == "" {
		matchValue = `([^\'"]*)`
	}
	attributeNameRegex, _ := regexp.Compile("[a-zA-Z][-.a-zA-Z0-9:_]*")
	spaceRx, _ := regexp.Compile("\\s*")
	tagAttrsRegex := spaceRx.ReplaceAllString(`
        (?:\s+
          `+attributeNameRegex.String()+`
          (?:\s*=\s*
            (?:
              '[^']*'
              |"[^"]*"
              |[^'"/>\s]+
            )
          )?
        )*
    `, "")
	rx := "<" + tag + tagAttrsRegex
	if tagClass != "" {
		rx += `\s+class\s*=\s*['"]\s*(?:[^'" ]+\s+)*` + tagClass + `(?:\s+[^'" ]+)*\s*['"]`
		rx += tagAttrsRegex
	}
	if matchParam != "" {
		rx += `\s+` + matchParam + `\s*=\s*['"]` + matchValue + `['"]`
		rx += tagAttrsRegex
	}
	rx += "\\s*>"
	return rx
}
