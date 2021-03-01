package source

import (
	"fmt"
	"github.com/PaesslerAG/jsonpath"
	"github.com/PuerkitoBio/goquery"
)

// region Scrape Items collection

type JSONScrapeItems struct {
	Items []interface{}
}

func (j *JSONScrapeItems) Length() int {
	return len(j.Items)
}

func (j *JSONScrapeItems) Get(i int) RawScrapeItem {
	return &JSONScrapeItem{item: j.Items[i]}
}

type DomScrapeItems struct {
	Items *goquery.Selection
}

func NewDOMScrapeItems(s *goquery.Selection) *DomScrapeItems {
	return &DomScrapeItems{Items: s}
}

func (d *DomScrapeItems) Length() int {
	return d.Items.Length()
}

func (d *DomScrapeItems) Get(i int) RawScrapeItem {
	return &DomScrapeItem{Selection: d.Items.Eq(i)}
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

func (j *JSONScrapeItem) PrevAllFiltered(_ string) RawScrapeItem {
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

func (j *JSONScrapeItem) Attr(_ string) (string, bool) {
	return "", false
}

func (j *JSONScrapeItem) Remove() RawScrapeItem {
	panic("implement me")
}

func (j *JSONScrapeItem) FindWithSelector(block *SelectorBlock) RawScrapeItem {
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
	Selection *goquery.Selection
}

func NewDOMScrapeItem(DOM *goquery.Document) *DomScrapeItem {
	return &DomScrapeItem{DOM.First()}
}

func (d *DomScrapeItem) FindWithSelector(block * SelectorBlock) RawScrapeItem {
	return &DomScrapeItem{Selection: d.Selection.Find(block.Selector)}
}

func (d *DomScrapeItem) Find(pathOrSelector string) RawScrapeItem {
	return &DomScrapeItem{Selection: d.Selection.Find(pathOrSelector)}
}

// Text gets the combined text contents of each element in the set of matched
// elements, including their descendants.
func (d *DomScrapeItem) Text() string {
	return d.Selection.Text()
}

// PrevAllFiltered gets all the preceding siblings of each element in the
// Selection filtered by a selector. It returns a new Selection object
// containing the matched elements.
func (d *DomScrapeItem) PrevAllFiltered(selector string) RawScrapeItem {
	return &DomScrapeItem{d.Selection.PrevAllFiltered(selector)}
}

func (d *DomScrapeItem) First() RawScrapeItem {
	return &DomScrapeItem{d.Selection.First()}
}

func (d *DomScrapeItem) Length() int {
	return d.Selection.Length()
}

func (d *DomScrapeItem) Has(selector string) RawScrapeItem {
	return &DomScrapeItem{Selection: d.Selection.Has(selector)}
}

func (d *DomScrapeItem) Map(f func(int, RawScrapeItem) string) []string {
	return d.Selection.Map(func(i int, selection *goquery.Selection) string {
		r := &DomScrapeItem{Selection: selection}
		return f(i, r)
	})
}

// Is checks the current matched set of elements against a selector and
// returns true if at least one of these elements matches.
func (d *DomScrapeItem) Is(selector string) bool {
	return d.Selection.Is(selector)
}

// Attr gets the specified attribute's value for the first element in the
// Selection. To get the value for each element individually, use a looping
// construct such as Each or Map method.
func (d *DomScrapeItem) Attr(name string) (string, bool) {
	return d.Selection.Attr(name)
}

func (d *DomScrapeItem) Remove() RawScrapeItem {
	return &DomScrapeItem{d.Selection.Remove()}
}

// endregion

