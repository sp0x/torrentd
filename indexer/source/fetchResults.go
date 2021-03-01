package source

import (
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type HTTPResult struct {
	contentType string
	encoding    string
	Response    *http.Response
	StatusCode  int
}

func (fr *HTTPResult) URL() *url.URL{
	loc, _ := fr.Response.Location()
	return loc
}

func (fr *HTTPResult) ContentType() string {
	return fr.contentType
}

func (fr *HTTPResult) Encoding() string {
	return fr.encoding
}

func (fr *HTTPResult) Find(selector string) RawScrapeItems {
	return nil
}

type HTMLFetchResult struct {
	HTTPResult
	DOM *goquery.Document
}

func (h *HTMLFetchResult) Find(selector string) RawScrapeItems {
	return NewDOMScrapeItems(h.DOM.Find(selector))
}

type JSONFetchResult struct {
	HTTPResult
	Body []byte
}
