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

// WebClient is a content fetcher that deals with the state of sources
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
func (w *WebClient) Fetch(target *RequestOptions) (FetchResult, error) {
	if target == nil {
		return nil, errors.New("target is required for searching")
	}
	defer func() {
		// After we're done we'll cleanup the history of the browser.
		w.Cleanup()
	}()
	var err error
	var result FetchResult
	switch target.Method {
	case "", searchMethodGet:
		if len(target.Values) > 0 {
			pURL, err := url.Parse(fmt.Sprintf("%s?%s", target.URL, target.Values.Encode()))
			if err != nil {
				return nil, err
			}
			target.URL = pURL
		}
		if err = w.get(target); err != nil {
			if w.errorHandler != nil {
				w.errorHandler(target)
			}
			return nil, err
		}
		result = extractResponseResult(w.Browser)
	case searchMethodPost:
		result, err = w.Post(target)
		if err != nil {
			if w.errorHandler != nil {
				w.errorHandler(target)
			}
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown search method %q", target.Method)
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

func (w *WebClient) get(reqOptions *RequestOptions) error {
	w.applyOptions(reqOptions)
	err := w.Browser.Open(reqOptions.URL.String())
	if err != nil {
		return err
	}
	if w.Cacher != nil && w.Cacher.IsCacheable() {
		_ = w.Cacher.CachePage(w.Browser.NewTab())
	}

	log.
		WithFields(log.Fields{"code": w.Browser.StatusCode(), "page": w.Browser.Url()}).
		Debugf("Finished request")
	if err = w.handleMetaRefreshHeader(reqOptions); err != nil {
		if w.errorHandler != nil {
			w.errorHandler(reqOptions)
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
