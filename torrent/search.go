package torrent

import "github.com/PuerkitoBio/goquery"

type Search struct {
	doc *goquery.Document
	id  string
}
