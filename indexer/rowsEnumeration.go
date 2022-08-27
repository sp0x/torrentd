package indexer

import (
	"encoding/json"
	"errors"

	"github.com/PuerkitoBio/goquery"

	"github.com/sp0x/torrentd/indexer/source"
)

func (r *Runner) extractScrapeItems(result source.FetchResult, srch *workerJob) (source.RawScrapeItems, error) {
	switch value := result.(type) {
	case *source.HTMLFetchResult:
		return r.extractItemsFromDom(value.DOM.First(), srch)
	case *source.JSONFetchResult:
		return r.extractItemsFromJSON(value.Body)
	}
	return nil, nil
}

func (r *Runner) extractItemsFromJSON(body []byte) (*source.JSONScrapeItems, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	node := data[r.definition.Search.Rows.Path]
	items := node.([]interface{})
	return &source.JSONScrapeItems{
		Items: items,
	}, nil
}

func (r *Runner) extractItemsFromDom(dom *goquery.Selection, srch *workerJob) (*source.DomScrapeItems, error) {
	if dom == nil {
		return nil, errors.New("DOM was nil")
	}
	updateSearchDataFromScrapeItem(r, srch, &source.DomScrapeItem{Selection: dom.First()})
	// merge following rows for After selector
	err := r.clearDom(dom)
	if err != nil {
		return nil, err
	}
	rows := dom.Find(r.definition.Search.Rows.Selector)
	return &source.DomScrapeItems{
		Items: rows,
	}, nil
}
