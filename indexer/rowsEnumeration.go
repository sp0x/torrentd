package indexer

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PaesslerAG/jsonpath"
	"github.com/PuerkitoBio/goquery"
	"github.com/sp0x/torrentd/indexer/source"
	"github.com/sp0x/torrentd/indexer/source/web"
)

type RawScrapeItems interface {
	Length() int
	Get(i int) RawScrapeItem
}

type RawScrapeItem interface {
	// FindWithSelector finds a child element using a selector block
	FindWithSelector(block *selectorBlock) RawScrapeItem
	// Find a child element using a selector or path.
	Find(selectorOrPath string) RawScrapeItem
	// Length of the child elements
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
	// Text gets the combined text contents of each element in the set of matched
	// elements, including their descendants.
	Text() string
	// Attr gets the specified attribute's value for the first element in the
	// Selection. To get the value for each element individually, use a looping
	// construct such as Each or Map method.
	Attr(attributeName string) (string, bool)
	// Remove removes the set of matched elements from the document.
	// It returns the same selection, now consisting of nodes not in the document.
	Remove() RawScrapeItem
	// PrevAllFiltered gets all the preceding siblings of each element in the
	// Selection filtered by a selector. It returns a new Selection object
	// containing the matched elements.
	PrevAllFiltered(selector string) RawScrapeItem
	First() RawScrapeItem
}

// region Scrape items collection

type JSONScrapeItems struct {
	items []interface{}
}

func (j *JSONScrapeItems) Length() int {
	return len(j.items)
}

func (j *JSONScrapeItems) Get(i int) RawScrapeItem {
	return &JSONScrapeItem{item: j.items[i]}
}

type DomScrapeItems struct {
	items *goquery.Selection
}

func (d *DomScrapeItems) Length() int {
	return d.items.Length()
}

func (d *DomScrapeItems) Get(i int) RawScrapeItem {
	return &DomScrapeItem{selection: d.items.Eq(i)}
}

// endregion

// region Scrape item

type JSONScrapeItem struct {
	item interface{}
}

func (j *JSONScrapeItem) First() RawScrapeItem {
	switch value := j.item.(type) {
	case []interface{}:
		return &JSONScrapeItem{value[0]}
	default:
		return j
	}
}

func (j *JSONScrapeItem) PrevAllFiltered(selector string) RawScrapeItem {
	panic("implement me")
}

func (j *JSONScrapeItem) Find(selectorOrPath string) RawScrapeItem {
	match, err := jsonpath.Get(selectorOrPath, j.item)
	if err != nil {
		return &JSONScrapeItem{item: nil}
	}
	return &JSONScrapeItem{item: match}
}

func (j *JSONScrapeItem) Has(selector string) RawScrapeItem {
	match, err := jsonpath.Get(selector, j.item)
	if match == nil || err != nil {
		return nil
	}
	return &JSONScrapeItem{match}
}

func (j *JSONScrapeItem) Map(f func(int, RawScrapeItem) string) []string {
	switch value := j.item.(type) {
	case []interface{}:
		output := make([]string, len(value))
		for i, item := range value {
			output[i] = f(i, &JSONScrapeItem{item})
		}
		return output
	default:
		return []string{f(0, &JSONScrapeItem{j.item})}
	}
}

func (j *JSONScrapeItem) Text() string {
	return fmt.Sprint(j.item)
}

func (j *JSONScrapeItem) Attr(attributeName string) (string, bool) {
	return "", false
}

func (j *JSONScrapeItem) Remove() RawScrapeItem {
	panic("implement me")
}

func (j *JSONScrapeItem) FindWithSelector(block *selectorBlock) RawScrapeItem {
	match, err := jsonpath.Get(block.Path, j.item)
	if err != nil {
		return &JSONScrapeItem{item: nil}
	}
	return &JSONScrapeItem{item: match}
}

func (j *JSONScrapeItem) Length() int {
	switch value := j.item.(type) {
	case []interface{}:
		return len(value)
	default:
		return 1
	}
}

// Is checks the current matched set of elements against a selector and
// returns true if at least one of these elements matches.
func (j *JSONScrapeItem) Is(_ string) bool {
	return false
}

type DomScrapeItem struct {
	selection *goquery.Selection
}

func (d *DomScrapeItem) FindWithSelector(block *selectorBlock) RawScrapeItem {
	return &DomScrapeItem{selection: d.selection.Find(block.Selector)}
}

func (d *DomScrapeItem) Find(pathOrSelector string) RawScrapeItem {
	return &DomScrapeItem{selection: d.selection.Find(pathOrSelector)}
}

// Text gets the combined text contents of each element in the set of matched
// elements, including their descendants.
func (d *DomScrapeItem) Text() string {
	return d.selection.Text()
}

// PrevAllFiltered gets all the preceding siblings of each element in the
// Selection filtered by a selector. It returns a new Selection object
// containing the matched elements.
func (d *DomScrapeItem) PrevAllFiltered(selector string) RawScrapeItem {
	return &DomScrapeItem{d.selection.PrevAllFiltered(selector)}
}

func (d *DomScrapeItem) First() RawScrapeItem {
	return &DomScrapeItem{d.selection.First()}
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

// Attr gets the specified attribute's value for the first element in the
// Selection. To get the value for each element individually, use a looping
// construct such as Each or Map method.
func (d *DomScrapeItem) Attr(name string) (string, bool) {
	return d.selection.Attr(name)
}

func (d *DomScrapeItem) Remove() RawScrapeItem {
	return &DomScrapeItem{d.selection.Remove()}
}

// endregion

func (r *Runner) getRows(result source.FetchResult, runCtx *RunContext) (RawScrapeItems, error) {
	switch value := result.(type) {
	case *web.HTMLFetchResult:
		return r.getRowsFromDom(value.Dom.First(), runCtx)
	case *web.JSONFetchResult:
		return r.getRowsFromJSON(value.Body)
	}
	return nil, nil
}

func (r *Runner) getRowsFromJSON(body []byte) (*JSONScrapeItems, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}
	node := data[r.definition.Search.Rows.Path]
	items := node.([]interface{})
	return &JSONScrapeItems{
		items: items,
	}, nil
}

func (r *Runner) getRowsFromDom(dom *goquery.Selection, runCtx *RunContext) (*DomScrapeItems, error) {
	if dom == nil {
		return nil, errors.New("DOM was nil")
	}
	setupContext(r, runCtx, &DomScrapeItem{dom.First()})
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
