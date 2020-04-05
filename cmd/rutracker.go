package main

import (
	"encoding/ascii85"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/requests"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"time"
)

type Rutracker struct {
	activeSearchId *int
	client         *http.Client
	loggedIn       bool
	lastRequest    time.Time
}

func newRuTracker() *Rutracker {
	rt := Rutracker{}
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

func (r *Rutracker) search(page int) (string, error) {
	if !r.loggedIn {
		return "", errors.New("not logged in")
	}
	if r.activeSearchId == nil {
		r.startSearch()
	}
	return "", nil
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

func (r *Rutracker) request(urlx string, data []byte) ([]byte, error) {
	maxPerSecond := 1
	minDiff := 1.0 / maxPerSecond
	timeElapsed := time.Now().Sub(r.lastRequest)
	if int(timeElapsed.Seconds()) < int(minDiff) {
		t := r.lastRequest.Add(time.Second * time.Duration(minDiff)).Sub(time.Now())
		time.Sleep(t)
	}
	requests.Get(r.client, "")
	r.lastRequest = time.Now()
	if data != nil {
		dascii := make([]byte, len(data), len(data))
		ascii85.Encode(dascii, data)
		data = dascii
	}
	resp, err := requests.Post(r.client, urlx, data, nil)
	if err != nil {
		return nil, err
	}
	resp = DecodeWindows1251(resp)
	return resp, nil
}

func (r *Rutracker) startSearch() {
	data := "prev_my=0&prev_new=0&prev_oop=0&o=1&s=2&tm=-1&pn=&nm=&submit=%CF%EE%E8%F1%EA"
	for _, forumId := range []int{
		46, 56, 98, 103, 249, 314, 500, 552, 709, 1260, 2076, 2123, 2139,
	} {
		data += "&f%5B%5D=" + strconv.Itoa(forumId)
	}
	page, err := r.request("https://rutracker.org/forum/tracker.php", []byte(data))
	if err != nil {
		return
	}
	rx, _ := regexp.Compile(tagRegex("a", "pg", "href", "tracker.php\\?search_id=([^&]+)[^'\"]*?"))
	matches := rx.FindAllStringSubmatch(string(page), -1)
	log.Print(matches)
}

func getNewTorrents(user, pass string) {
	log.Info("Searching for new torrents")

	tr1 := tagRegex("a", "tLink", "href", `viewtopic\.php\?t=(\d+)`)
	tr2 := tagRegex("td", "row4", "data-ts_text", "(\\d+)")
	torrentRegex, _ := regexp.Compile(tr1 + "(.+?)<a/>" +
		`.+` +
		tr2 +
		"\\s*<p>\\s*\\d{1,2}-(?:Янв|Фев|Мар|Апр|Май|Июн|Июл|Авг|Сен|Окт|Ноя|Дек)-\\d{2}</p>")
	client := newRuTracker()
	err := client.login(user, pass)
	if err != nil {
		log.Error("Could not log in.")
		return
	}

	pagesToScan := 1
	torrentsPerPage := 50
	totalTorrents := pagesToScan * torrentsPerPage
	page := 0
	log.Print(totalTorrents)
	for page = 0; page < pagesToScan; page++ {
		log.Infof("Getting page %d", page)
		pageContent, err := client.search(page)
		if err != nil {
			continue
		}
		matches := torrentRegex.FindAllString(pageContent, -1)
		log.Print(matches)
	}

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
