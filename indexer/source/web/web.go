package web

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/torrentd/indexer/source"
	"net/url"
	"regexp"
	"strings"
)

const (
	searchMethodPost = "post"
	searchMethodGet  = "get"
)

type ContentFetcher struct {
	Browser browser.Browsable
	Cacher  ContentCacher
}

type ContentCacher interface {
	CachePage(browsable browser.Browsable) error
}

func (w *ContentFetcher) Cleanup() {
	w.Browser.HistoryJar().Clear()
}

func (w *ContentFetcher) FetchUrl(url string) error {
	target := source.SearchTarget{Url: url}
	return w.get(target.Url)
}

//Gets the content from which we'll extract the search results
func (w *ContentFetcher) Fetch(target *source.SearchTarget) error {
	if target == nil {
		return errors.New("target is required for searching")
	}
	defer func() {
		//After we're done we'll cleanup the history of the browser.
		w.Cleanup()
	}()
	var err error
	switch target.Method {
	case "", searchMethodGet:
		if len(target.Values) > 0 {
			target.Url = fmt.Sprintf("%s?%s", target.Url, target.Values.Encode())
		}
		if err = w.get(target.Url); err != nil {
			return err
		}
	case searchMethodPost:
		if err = w.Post(target.Url, target.Values, true); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown search method %q", target.Method)
	}
	return nil
}

func (w *ContentFetcher) get(targetUrl string) error {
	logrus.WithField("target", targetUrl).
		Debug("Opening page")
	err := w.Browser.Open(targetUrl)
	if err != nil {
		return err
	}
	if w.Cacher != nil {
		_ = w.Cacher.CachePage(w.Browser.NewTab())
	}

	logrus.
		WithFields(logrus.Fields{"code": w.Browser.StatusCode(), "page": w.Browser.Url()}).
		Debugf("Finished request")
	if err = w.handleMetaRefreshHeader(); err != nil {
		return err
	}
	return nil
}

func (w *ContentFetcher) Post(url string, data url.Values, log bool) error {
	if log {
		logrus.
			WithFields(logrus.Fields{"url": url, "vals": data.Encode()}).
			Debugf("Posting to page")
	}
	if err := w.Browser.PostForm(url, data); err != nil {
		return err
	}
	if w.Cacher != nil {
		_ = w.Cacher.CachePage(w.Browser.NewTab())
	}
	logrus.
		WithFields(logrus.Fields{"code": w.Browser.StatusCode(), "page": w.Browser.Url()}).
		Debugf("Finished request")

	if err := w.handleMetaRefreshHeader(); err != nil {
		return err
	}
	return nil
}

// this should eventually upstream into surf browser
//Handle a header like: Refresh: 0;url=my_view_page.php
func (w *ContentFetcher) handleMetaRefreshHeader() error {
	h := w.Browser.ResponseHeaders()
	if refresh := h.Get("Refresh"); refresh != "" {
		requestUrl := w.Browser.State().Request.URL
		if s := regexp.MustCompile(`\s*;\s*`).Split(refresh, 2); len(s) == 2 {
			logrus.
				WithField("fields", s).
				Debug("Found refresh header")
			requestUrl.Path = strings.TrimPrefix(s[1], "url=")
			return w.get(requestUrl.String())
		}
	}
	return nil
}
