package web

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/source"
	"io"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	searchMethodPost = "post"
	searchMethodGet  = "get"
)

//ContentFetcher is a content fetcher that deals with the state of sources
type ContentFetcher struct {
	Browser            browser.Browsable
	Cacher             ContentCacher
	ConnectivityTester cache.ConnectivityTester
	options            FetchOptions
}

type FetchOptions struct {
	DumpData bool
}

func NewWebContentFetcher(browser browser.Browsable, contentCache ContentCacher, connectivityTester cache.ConnectivityTester, options FetchOptions) source.ContentFetcher {
	if connectivityTester == nil {
		panic("a connectivity tester is required")
	}
	return &ContentFetcher{
		Browser: browser,
		//We'll use the indexer to cache content.
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

func (w *ContentFetcher) FetchUrl(url string) error {
	target := source.SearchTarget{Url: url}
	err := w.get(target.Url)
	if err != nil {
		w.ConnectivityTester.Invalidate(target.Url)
	}
	return err
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
			w.ConnectivityTester.Invalidate(target.Url)
			return err
		}
	case searchMethodPost:
		if err = w.Post(target.Url, target.Values, true); err != nil {
			w.ConnectivityTester.Invalidate(target.Url)
			return err
		}

	default:
		return fmt.Errorf("unknown search method %q", target.Method)
	}
	w.postProcessData()
	return nil
}

func (w *ContentFetcher) postProcessData() {
	if !w.options.DumpData {
		return
	}
	//todo: dump data
	browserState := w.Browser.State()
	request := browserState.Request
	dirPath := path.Join("dumps", request.Host)
	dirPath = path.Join(dirPath,
		strings.Replace(fmt.Sprintf("%s_%s", request.Method, request.URL.Path), "/", "_", -1))
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 007)
	}
	timeNow := fmt.Sprintf("%d", time.Now().Unix())
	responseBodyFName := fmt.Sprintf("%s_resp.%s", timeNow, resolveResponseDumpFormat(browserState))
	requestExtension := resolveRequestDumpFormat(browserState)
	requestBodyFName := fmt.Sprintf("%s_req.%s", timeNow, requestExtension)

	responseBodyPath := path.Join(dirPath, responseBodyFName)
	requestBodyPath := path.Join(dirPath, requestBodyFName)

	responseFileWriter, err := os.Create(responseBodyPath)
	if err != nil {
		logrus.Warnf("could not dump response %s", responseBodyPath)
		return
	}
	//use browser url
	_, err = w.Browser.Download(responseFileWriter)
	if requestExtension != "" {
		requestFileWriter, _ := os.Create(requestBodyPath)
		getBody := browserState.Request.GetBody
		copyBody, _ := getBody()
		n, err := io.Copy(requestFileWriter, copyBody)
		if err != nil {
			logrus.Warnf("could not dump request body %s", requestBodyPath)
		} else {
			logrus.Debugf("written body with size %d bytes", n)
		}
	}

}

func resolveRequestDumpFormat(state *jar.State) string {
	if state.Request.Method == "GET" {
		return ""
	}
	contentType := state.Request.Header.Get("Content-Type")
	if contentType == "" {
		return ""
	}
	return contentTypeToFileExtension(contentType)
}

func resolveResponseDumpFormat(state *jar.State) string {
	contentType := state.Response.Header.Get("Content-Type")
	if contentType == "" {
		return "html"
	}
	return contentTypeToFileExtension(contentType)
}

func contentTypeToFileExtension(contentType string) string {
	switch contentType {
	case "application/json":
		return "json"
	}
	return "html"
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
		w.ConnectivityTester.Invalidate(targetUrl)
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
		w.ConnectivityTester.Invalidate(url)
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
			err := w.get(requestUrl.String())
			if err != nil {
				w.ConnectivityTester.Invalidate(requestUrl.String())
			}
			return err
		}
	}
	return nil
}
