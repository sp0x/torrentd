package source

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

type HTTPResult struct {
	contentType string
	encoding    string
	Response    *http.Response
	StatusCode  int
}

func (fr *HTTPResult) ContentType() string {
	return fr.contentType
}

func (fr *HTTPResult) Encoding() string {
	return fr.encoding
}

type HTMLFetchResult struct {
	HTTPResult
	DOM *goquery.Document
}

type JSONFetchResult struct {
	HTTPResult
	Body []byte
}
