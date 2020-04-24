package torrent

import (
	"github.com/sp0x/rutracker-rss/indexer/search"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Tracker interface {
	Login(username, password string) error
	GetTorrentLink(t *search.ExternalResultItem) string
	GetTorrentDownloadLink(t *search.ExternalResultItem) string
	GetDefaultOptions() *GenericSearchOptions
}

type BasicTracker struct {
	lastRequest time.Time
	//client          *http.Client
	FetchDefinition bool
}

//Send a request to the tracker
//func (r *BasicTracker) request(urlx string, data []byte, headers map[string]string) ([]byte, error) {
//	maxPerSecond := 1
//	minDiff := 1.0 / maxPerSecond
//	timeElapsed := time.Now().Sub(r.lastRequest)
//	if int(timeElapsed.Seconds()) < int(minDiff) {
//		t := r.lastRequest.Add(time.Second * time.Duration(minDiff)).Sub(time.Now())
//		time.Sleep(t)
//	}
//	r.lastRequest = time.Now()
//	log.Debugf("GET %s\n", urlx)
//	log.Debugf("DATA: %v\n", string(data))
//	resp, err := requests.Get(r.client, urlx, headers)
//	if err != nil {
//		return nil, err
//	}
//	decoder := encoding.GetEncoding("windows1251").NewDecoder()
//	resp, _ = decoder.Bytes(resp)
//	return resp, nil
//}

func newWebClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		DisableCompression: false,
	}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
		Jar:       jar,
	}
	return client
}
