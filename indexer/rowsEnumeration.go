package indexer

import (
	"encoding/json"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/sp0x/torrentd/indexer/source"
	"github.com/sp0x/torrentd/indexer/source/web"
)

type RawScrapeItems interface {
	Length() int
	Get(i int) RawScrapeItem
}

type RawScrapeItem interface {
	Find(block *selectorBlock) RawScrapeItem
	Length() int

	// Is checks the current matched set of elements against a selector and
	// returns true if at least one of these elements matches.
	Is(selector string) bool
	// Has reduces the set of matched elements to those that have a descendant
	// that matches the selector.
	// It returns a new Selection object with the matching elements.
	Has(selector string) RawScrapeItem
	// Map passes each element in the current matched set through a function,
	// producing a slice of string holding the returned values. The function
	// f is called for each element in the selection with the index of the
	// element in that selection starting at 0, and a *Selection that contains
	// only that element.
	Map(f func(int, RawScrapeItem) string) []string
}

type JsonScrapeItems struct {
	items []interface{}
}

type DomScrapeItems struct {
	items *goquery.Selection
}

func (j *JsonScrapeItems) Length() int {
	return len(j.items)
}

func (j *JsonScrapeItems) Get(i int) RawScrapeItem {
	return &JsonScrapeItem{item: j.items[i]}
}

func (d *DomScrapeItems) Length() int {
	return d.items.Length()
}

func (d *DomScrapeItems) Get(i int) RawScrapeItem {
	return &DomScrapeItem{selection: d.items.Eq(i)}
}

//region Scrape item
type DomScrapeItem struct {
	selection *goquery.Selection
}
type JsonScrapeItem struct {
	item interface{}
}

func (d *DomScrapeItem) Find(block *selectorBlock) RawScrapeItem {
	return &DomScrapeItem{selection: d.selection.Find(block.Selector)}
}

func (d *DomScrapeItem) Length() int {
	return d.selection.Length()
}

func (d *DomScrapeItem) Has(selector string) RawScrapeItem {
	return &DomScrapeItem{selection: d.selection.Has(selector)}
}

func (d *DomScrapeItem) Map(f func(int, RawScrapeItem) string) []string {
	return d.selection.Map(func(i int, selection *goquery.Selection) string {
		r := &DomScrapeItem{selection: selection}
		return f(i, r)
	})
}

// Is checks the current matched set of elements against a selector and
// returns true if at least one of these elements matches.
func (d *DomScrapeItem) Is(selector string) bool {
	return d.selection.Is(selector)
}

func (j *JsonScrapeItem) Find(block *selectorBlock) RawScrapeItem {
	return &JsonScrapeItem{item: j.item[block.Path]}
}

func (j *JsonScrapeItem) Length() int {
	return 1
}

// Is checks the current matched set of elements against a selector and
// returns true if at least one of these elements matches.
func (j *JsonScrapeItem) Is(_ string) bool {
	return false
}

//endregion

func (r *Runner) getRows(result source.FetchResult, runCtx *RunContext) (RawScrapeItems, error) {
	switch value := result.(type) {
	case *web.HtmlFetchResult:
		return r.getRowsFromDom(value.Dom.First(), runCtx)
	case *web.JsonFetchResult:
		return r.getRowsFromJson(value.Body)
	}
	return nil, nil
}

func (r *Runner) getRowsFromJson(body []byte) (*JsonScrapeItems, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	node := data[r.definition.Search.Rows.Path]
	items := node.([]interface{})
	return &JsonScrapeItems{
		items: items,
	}, nil
}

func (r *Runner) getRowsFromDom(dom *goquery.Selection, runCtx *RunContext) (*DomScrapeItems, error) {
	if dom == nil {
		return nil, errors.New("DOM was nil")
	}
	setupContext(r, runCtx, dom)
	// merge following rows for After selector
	err := r.clearDom(dom)
	if err != nil {
		return nil, err
	}
	rows := dom.Find(r.definition.Search.Rows.Selector)
	return &DomScrapeItems{
		items: rows,
	}, nil
}
