package torrent

import (
	"github.com/sp0x/rutracker-rss/db"
	"github.com/sp0x/rutracker-rss/requests"
	"net/http"
	"time"
)

type Tracker interface {
	Login(username, password string) error
	GetTorrentLink(t *db.Torrent) string
	GetTorrentDownloadLink(t *db.Torrent) string
	GetDefaultOptions() *FetchOptions
}

type BasicTracker struct {
	lastRequest     time.Time
	client          *http.Client
	storage         Storage
	FetchDefinition bool
}

//Send a request to the tracker
func (r *BasicTracker) request(urlx string, data []byte, headers map[string]string) ([]byte, error) {
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