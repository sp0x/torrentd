package search

import "github.com/PuerkitoBio/goquery"

type Search struct {
	DOM         *goquery.Selection
	Id          string
	CurrentPage int
	StartIndex  int
	Results     []ResultItem
}

func (s *Search) GetDocument() *goquery.Selection {
	return s.DOM
}
