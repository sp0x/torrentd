package torrent

import "github.com/PuerkitoBio/goquery"

type Search struct {
	doc *goquery.Document
	id  string
}

func (s *Search) GetDocument() *goquery.Document {
	return s.doc
}
