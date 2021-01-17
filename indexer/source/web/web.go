package web

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"

	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/source"
)

const (
	searchMethodPost = "post"
	searchMethodGet  = "get"
)

// ContentFetcher is a content fetcher that deals with the state of sources
type ContentFetcher struct {
	Browser            browser.Browsable
	Cacher             ContentCacher
	ConnectivityTester cache.ConnectivityTester
	options            FetchOptions
}

type FetchOptions struct {
	ShouldDumpData bool
	FakeReferer    bool
}

func NewWebContentFetcher(browser browser.Browsable,
	contentCache ContentCacher,
	connectivityTester cache.ConnectivityTester,
	options FetchOptions) source.ContentFetcher {
	if connectivityTester == nil {
		panic("a connectivity tester is required")
	}
	return &ContentFetcher{
		Browser: browser,
		// We'll use the indexer to cache content.
		Cacher:             contentCache,
		ConnectivityTester: connectivityTester,
		options:            options,
	}
}

type ContentCacher interface {
	CachePage(browsable browser.Browsable) error
}

func (w *ContentFetcher) Cleanup() {
	w.Browser.HistoryJar().Clear()
}

func (w *ContentFetcher) FetchURL(url string) error {
	target := source.SearchTarget{URL: url}
	err := w.get(target.URL)
	if err != nil {
		w.ConnectivityTester.Invalidate(target.URL)
	}
	return err
}

// Gets the content from which we'll extract the search results
func (w *ContentFetcher) Fetch(target *source.SearchTarget) (source.FetchResult, error) {
	if target == nil {
		return nil, errors.New("target is required for searching")
	}
	defer func() {
		// After we're done we'll cleanup the history of the browser.
		w.Cleanup()
	}()
	var err error
	switch target.Method {
	case "", searchMethodGet:
		if len(target.Values) > 0 {
			target.URL = fmt.Sprintf("%s?%s", target.URL, target.Values.Encode())
		}
		if err = w.get(target.URL); err != nil {
			w.ConnectivityTester.Invalidate(target.URL)
			return nil, err
		}
	case searchMethodPost:
		if err = w.Post(target.URL, target.Values, true); err != nil {
			w.ConnectivityTester.Invalidate(target.URL)
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown search method %q", target.Method)
	}
	w.dumpFetchData()

	return extractResponseResult(w.Browser), nil
}

func extractResponseResult(browser browser.Browsable) source.FetchResult {
	state := browser.State()
	if state.Response == nil {
		return &HTTPResult{}
	}
	fqContentType := state.Response.Header.Get("content-type")
	contentSplit := strings.Split(fqContentType, ";")
	contentEncoding := "utf8"
	if len(contentSplit) > 1 {
		contentEncoding = contentSplit[1]
	}
	rootFetchResult := HTTPResult{
		contentType: contentSplit[0],
		encoding:    contentEncoding,
		Response:    state.Response,
	}

	if contentSplit[0] == "application/json" {
		return &JSONFetchResult{
			HTTPResult: rootFetchResult,
			Body:       browser.RawBody(),
		}
	}

	return &HTMLFetchResult{
		HTTPResult: rootFetchResult,
		Dom:        state.Dom,
	}
}

func (w *ContentFetcher) get(targetURL string) error {
	logrus.WithField("target", targetURL).
		Debug("Opening page")
	err := w.Browser.Open(targetURL)
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
		w.ConnectivityTester.Invalidate(targetURL)
		return err
	}
	return nil
}

func (w *ContentFetcher) Post(urlStr string, data url.Values, log bool) error {
	if log {
		logrus.
			WithFields(logrus.Fields{"urlStr": urlStr, "vals": data.Encode()}).
			Debugf("Posting to page")
	}
	if w.options.FakeReferer {
		w.fakeBrowserReferer(urlStr)
	}
	if err := w.Browser.PostForm(urlStr, data); err != nil {
		return err
	}
	if w.Cacher != nil {
		_ = w.Cacher.CachePage(w.Browser.NewTab())
	}
	logrus.
		WithFields(logrus.Fields{"code": w.Browser.StatusCode(), "page": w.Browser.Url()}).
		Debugf("Finished request")

	if err := w.handleMetaRefreshHeader(); err != nil {
		w.ConnectivityTester.Invalidate(urlStr)
		return err
	}
	w.dumpFetchData()
	return nil
}

func (w *ContentFetcher) fakeBrowserReferer(urlStr string) {
	state := w.Browser.State()
	refURL, _ := url.Parse(urlStr)
	if state.Request == nil {
		state.Request = &http.Request{}
	}
	state.Request.URL = refURL
	if state.Response != nil {
		state.Response.Request.URL = refURL
	}
}

// this should eventually upstream into surf browser
// Handle a header like: Refresh: 0;url=my_view_page.php
func (w *ContentFetcher) handleMetaRefreshHeader() error {
	h := w.Browser.ResponseHeaders()
	if refresh := h.Get("Refresh"); refresh != "" {
		requestURL := w.Browser.State().Request.URL
		if s := regexp.MustCompile(`\s*;\s*`).Split(refresh, 2); len(s) == 2 {
			logrus.
				WithField("fields", s).
				Debug("Found refresh header")
			requestURL.Path = strings.TrimPrefix(s[1], "url=")
			err := w.get(requestURL.String())
			if err != nil {
				w.ConnectivityTester.Invalidate(requestURL.String())
			}
			return err
		}
	}
	return nil
}
