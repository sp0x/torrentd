package source

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"
)

const (
	searchMethodPost = "post"
	searchMethodGet  = "get"
)

type WebClient struct {
	Browser      browser.Browsable
	Cacher       ContentCacher
	options      FetchOptions
	errorHandler func(options *RequestOptions)
}

func (w *WebClient) SetErrorHandler(callback func(options *RequestOptions)) {
	w.errorHandler = callback
}

func NewWebContentFetcher(browser browser.Browsable,
	contentCache ContentCacher,
	options FetchOptions) *WebClient {
	browser.SetCookieJar(jar.NewMemoryCookies())
	return &WebClient{
		Browser: browser,
		// We'll use the indexer to cache content.
		Cacher:  contentCache,
		options: options,
	}
}

type ContentCacher interface {
	CachePage(browsable browser.Browsable) error
	IsCacheable() bool
}

func (w *WebClient) Cleanup() {
	w.Browser.HistoryJar().Clear()
}

// Gets the content from which we'll extract the search results
func (w *WebClient) Fetch(req *RequestOptions) (FetchResult, error) {
	if req == nil {
		return nil, errors.New("req is required for searching")
	}
	defer func() {
		// After we're done we'll cleanup the history of the browser.
		w.Cleanup()
	}()
	var err error
	var result FetchResult
	switch req.Method {
	case "", searchMethodGet:
		if err = w.get(req); err != nil {
			if w.errorHandler != nil {
				w.errorHandler(req)
			}
			return nil, err
		}
		result = extractResponseResult(w.Browser)
	case searchMethodPost:
		postResult, err := w.Post(req)
		if err != nil {
			if w.errorHandler != nil {
				w.errorHandler(req)
			}
			return nil, err
		}
		result = postResult

	default:
		return nil, fmt.Errorf("unknown search method %q", req.Method)
	}
	w.dumpFetchData()

	return result, nil
}

func extractResponseResult(browser browser.Browsable) FetchResult {
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
		StatusCode:  state.Response.StatusCode,
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

func (w *WebClient) get(req *RequestOptions) error {
	destURL := req.URL
	if len(req.Values) > 0 {
		pURL, err := url.Parse(fmt.Sprintf("%s?%s", req.URL, req.Values.Encode()))
		if err != nil {
			return err
		}
		destURL = pURL
	}
	w.applyOptions(req)
	err := w.Browser.Open(destURL.String())
	if err != nil {
		return err
	}
	if w.Cacher != nil && w.Cacher.IsCacheable() {
		_ = w.Cacher.CachePage(w.Browser.NewTab())
	}

	if err = w.handleMetaRefreshHeader(req); err != nil {
		if w.errorHandler != nil {
			w.errorHandler(req)
		}
		return err
	}
	return nil
}

func (w *WebClient) applyOptions(reqOptions *RequestOptions) {
	referer := ""
	if w.options.FakeReferer && reqOptions.Referer == nil {
		referer = reqOptions.URL.String()
	} else if reqOptions.Referer != nil {
		referer = reqOptions.Referer.String()
	}

	if referer != "" {
		w.Browser.SetHeadersJar(http.Header{
			"referer": []string{referer},
		})
	}
	if reqOptions.CookieJar != nil {
		w.Browser.SetCookieJar(reqOptions.CookieJar)
	}
}

func (w *WebClient) URL() *url.URL {
	return w.Browser.Url()
}

func (w *WebClient) Clone() ContentFetcher {
	f := &WebClient{}
	*f = *w
	f.Browser = f.Browser.NewTab()
	return f
}

func (w *WebClient) Open(opts *RequestOptions) (FetchResult, error) {
	if opts.Encoding != "" || opts.NoEncoding {
		w.Browser.SetEncoding(opts.Encoding)
	}
	err := w.Browser.Open(opts.URL.String())
	if err != nil {
		return nil, err
	}
	return extractResponseResult(w.Browser), nil
}

func (w *WebClient) Download(buffer io.Writer) (int64, error) {
	return w.Browser.Download(buffer)
}

func (w *WebClient) Post(reqOps *RequestOptions) (FetchResult, error) {
	urlStr := reqOps.URL
	values := reqOps.Values

	w.applyOptions(reqOps)
	if err := w.Browser.PostForm(urlStr.String(), values); err != nil {
		return nil, err
	}
	if w.Cacher != nil {
		_ = w.Cacher.CachePage(w.Browser.NewTab())
	}

	if err := w.handleMetaRefreshHeader(reqOps); err != nil {
		return nil, err
	}
	w.dumpFetchData()
	return extractResponseResult(w.Browser), nil
}

// this should eventually upstream into surf browser
// Handle a header like: Refresh: 0;url=my_view_page.php
func (w *WebClient) handleMetaRefreshHeader(reqOptions *RequestOptions) error {
	h := w.Browser.ResponseHeaders()
	if refresh := h.Get("Refresh"); refresh != "" {
		requestURL := w.Browser.State().Request.URL
		if s := regexp.MustCompile(`\s*;\s*`).Split(refresh, 2); len(s) == 2 {
			log.
				WithField("fields", s).
				Info("Found refresh header")
			requestURL.Path = strings.TrimPrefix(s[1], "url=")
			reqOptions.URL = requestURL

			err := w.get(reqOptions)
			if err != nil {
				if w.errorHandler != nil {
					w.errorHandler(reqOptions)
				}
			}
			return err
		}
	}
	return nil
}
