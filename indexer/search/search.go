package search

import "github.com/PuerkitoBio/goquery"

type Search struct {
	DOM         *goquery.Document
	Id          string
	CurrentPage int
	StartIndex  int
	Results     []ResultItem
}

func (s *Search) GetDocument() *goquery.Document {
	return s.DOM
}
