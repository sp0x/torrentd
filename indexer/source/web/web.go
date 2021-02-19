package web

import (
	"errors"
	"fmt"
	"io"
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

// Fetcher is a content fetcher that deals with the state of sources
type Fetcher struct {
	Browser            browser.Browsable
	Cacher             ContentCacher
	ConnectivityTester cache.ConnectivityTester
	options            source.FetchOptions
}

func NewWebContentFetcher(browser browser.Browsable,
	contentCache ContentCacher,
	connectivityTester cache.ConnectivityTester,
	options source.FetchOptions) source.ContentFetcher {
	if connectivityTester == nil {
		panic("a connectivity tester is required")
	}
	return &Fetcher{
		Browser: browser,
		// We'll use the indexer to cache content.
		Cacher:             contentCache,
		ConnectivityTester: connectivityTester,
		options:            options,
	}
}

type ContentCacher interface {
	CachePage(browsable browser.Browsable) error
	IsCacheable() bool
}

func (w *Fetcher) Cleanup() {
	w.Browser.HistoryJar().Clear()
	w.ConnectivityTester.ClearBrowser()
}

//func (w *Fetcher) Get(url string) error {
//	target := source.FetchOptions{URL: url}
//	err := w.get(target.URL)
//	if err != nil {
//		w.ConnectivityTester.Invalidate(target.URL)
//	}
//	switch value := result.(type) {
//	case *web.HTMLFetchResult:
//		return r.getRowsFromDom(value.DOM.First(), runCtx)
//	case *web.JSONFetchResult:
//		return r.getRowsFromJSON(value.Body)
//	}
//	return err
//}

// Gets the content from which we'll extract the search results
func (w *Fetcher) Fetch(target *source.FetchOptions) (source.FetchResult, error) {
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
		DOM:        state.Dom,
	}
}

func (w *Fetcher) get(targetURL string) error {
	logrus.WithField("target", targetURL).
		Debug("Opening page")
	err := w.Browser.Open(targetURL)
	if err != nil {
		return err
	}
	if w.Cacher != nil && w.Cacher.IsCacheable() {
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

func (w *Fetcher) URL() *url.URL {
	browserUrl := w.Browser.Url()
	return browserUrl
}

func (w *Fetcher) Clone() source.ContentFetcher {
	f := &Fetcher{}
	*f = *w
	f.Browser = f.Browser.NewTab()
	return f
}

func (w *Fetcher) Open(opts *source.FetchOptions) error {
	if opts.Encoding != "" || opts.NoEncoding {
		w.Browser.SetEncoding(opts.Encoding)
	}
	return w.Browser.Open(opts.URL)
}

func (w *Fetcher) Download(buffer io.Writer) (int64, error) {
	return w.Browser.Download(buffer)
}

func (w *Fetcher) Post(urlStr string, data url.Values, log bool) error {
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

func (w *Fetcher) fakeBrowserReferer(urlStr string) {
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
func (w *Fetcher) handleMetaRefreshHeader() error {
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
