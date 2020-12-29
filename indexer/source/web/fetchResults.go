package web

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
)

type HttpResult struct {
	contentType string
	encoding    string
	Response    *http.Response
}

func (fr *HttpResult) ContentType() string {
	return fr.contentType
}

func (fr *HttpResult) Encoding() string {
	return fr.encoding
}

type HtmlFetchResult struct {
	HttpResult
	Dom *goquery.Document
}

type JsonFetchResult struct {
	HttpResult
	Body []byte
}
