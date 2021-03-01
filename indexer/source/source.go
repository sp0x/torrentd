package source

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/templates"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var filterService FilterService

type FilterProvider interface {
	Filter(fType string, args interface{}, value string) (string, error)
}

type RawScrapeItems interface {
	Length() int
	Get(i int) RawScrapeItem
}

type filterBlock struct {
	Name string      `yaml:"name"`
	Args interface{} `yaml:"args"`
}

type SelectorBlock struct {
	// The css selector to use to look for a match
	Selector string `yaml:"selector"`
	// The xml/json path to look for
	Path string `yaml:"path"`
	//
	Pattern      string            `yaml:"pattern"`
	TextVal      string            `yaml:"text"`
	Attribute    string            `yaml:"attribute,omitempty"`
	Remove       string            `yaml:"remove,omitempty"`
	Filters      []filterBlock     `yaml:"filters,omitempty"`
	Case         map[string]string `yaml:"case,omitempty"`
	FilterConfig map[string]string `yaml:"filterconfig"`
	// If we'll use all the values
	All bool `yaml:"all"`
}

type RawScrapeItem interface {
	// FindWithSelector finds a child element using a selector block
	FindWithSelector(block *SelectorBlock) RawScrapeItem
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
	// f is called for each element in the Selection with the index of the
	// element in that Selection starting at 0, and a *Selection that contains
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
	// It returns the same Selection, now consisting of nodes not in the document.
	Remove() RawScrapeItem
	// PrevAllFiltered gets all the preceding siblings of each element in the
	// Selection filtered by a selector. It returns a new Selection object
	// containing the matched elements.
	PrevAllFiltered(selector string) RawScrapeItem
	First() RawScrapeItem
}

type FetchOptions struct {
	ShouldDumpData bool
	FakeReferer    bool
}

type RequestOptions struct {
	URL        *url.URL
	Values     url.Values
	Method     string
	Encoding   string
	NoEncoding bool
	CookieJar  http.CookieJar
	Referer    *url.URL
}

func NewRequestOptions(destURL *url.URL) *RequestOptions {
	return &RequestOptions{
		URL: destURL,
	}
}

type FetchResult interface {
	ContentType() string
	Encoding() string
	Find(selector string) RawScrapeItems
}

//go:generate mockgen -source source.go -destination=mocks/source.go -package=mocks
type ContentFetcher interface {
	Cleanup()
	Fetch(target *RequestOptions) (FetchResult, error)
	Post(options *RequestOptions) (FetchResult, error)
	URL() *url.URL
	Clone() ContentFetcher
	Open(options *RequestOptions) (FetchResult, error)
	Download(buffer io.Writer) (int64, error)
	SetErrorHandler(callback func(options *RequestOptions))
}

func (s *SelectorBlock) IsMatching(selection *goquery.Selection) bool {
	return !s.IsEmpty() && (selection.Find(s.Selector).Length() > 0 || s.TextVal != "")
}

// Match using the selector to get the text of that element
func (s *SelectorBlock) Match(from RawScrapeItem) (interface{}, error) {
	if s.TextVal != "" {
		return s.TextVal, nil
	}
	if s.Selector != "" {
		result := from.FindWithSelector(s)
		if result.Length() == 0 {
			return "", fmt.Errorf("failed to match selector %q", s.Selector)
		}
		if s.All {
			return s.Texts(result)
		}
		return s.Text(result)
	}
	if s.Pattern != "" {
		return s.Pattern, nil
	}
	if s.Path != "" {
		result := from.Find(s.Path)
		if result.Length() == 0 {
			return "", fmt.Errorf("failed to match selector %q", s.Selector)
		}
		if s.All {
			return s.Texts(result)
		}
		return s.Text(result)
	}
	return s.Text(from)
}

func (s *SelectorBlock) MatchRawText(from RawScrapeItem) (string, error) {
	if s.TextVal != "" {
		return s.TextVal, nil
	}
	if s.Selector != "" {
		result := from.FindWithSelector(s)
		if result.Length() == 0 {
			return "", fmt.Errorf("failed to match selector %q", s.Selector)
		}
		return s.TextRaw(result)
	}
	if s.Pattern != "" {
		return s.Pattern, nil
	}
	return s.TextRaw(from)
}

func (s *SelectorBlock) TextRaw(el RawScrapeItem) (string, error) {
	if s.TextVal != "" {
		return s.TextVal, nil
	}
	if s.Remove != "" {
		el.Find(s.Remove).Remove()
	}
	if s.Case != nil {
		// filterLogger.WithFields(logrus.Fields{"case": s.Case}).
		//	Debugf("Applying case to Selection")
		for pattern, value := range s.Case {
			if el.Is(pattern) || el.Has(pattern).Length() >= 1 {
				return s.FilterText(value)
			}
		}
		return "", errors.New("none of the cases match")
	}
	output := strings.TrimSpace(el.Text())
	output = spaceRx.ReplaceAllString(output, " ")
	if s.Attribute != "" {
		val, exists := el.Attr(s.Attribute)
		if !exists {
			return "", fmt.Errorf("requested attribute %q doesn't exist", s.Attribute)
		}
		output = val
	}
	return output, nil
}

func (s *SelectorBlock) Texts(element RawScrapeItem) ([]string, error) {
	if s.TextVal != "" {
		filterResult, err := s.FilterText(s.TextVal)
		if err != nil {
			return nil, err
		}
		return []string{filterResult}, nil
	}
	if s.Remove != "" {
		element.Find(s.Remove).Remove()
	}
	if s.Case != nil {
		for pattern, value := range s.Case {
			if element.Is(pattern) || element.Has(pattern).Length() >= 1 {
				value, _ := s.FilterText(value)
				return []string{value}, nil
			}
		}
		return []string{}, errors.New("none of the cases match")
	}
	matches := element.Map(func(i int, selection RawScrapeItem) string {
		output := strings.TrimSpace(selection.Text())
		output = spaceRx.ReplaceAllString(output, " ")
		if s.Attribute != "" {
			val, exists := selection.Attr(s.Attribute)
			if !exists {
				return ""
			}
			output = val
		}
		filteredResult, err := s.FilterText(output)
		if err != nil {
			return ""
		}
		return filteredResult
	})
	return matches, nil
}

// Text extracts text from the Selection, applying all filters
func (s *SelectorBlock) Text(el RawScrapeItem) (string, error) {
	if s.TextVal != "" {
		return s.FilterText(s.TextVal)
	}

	if s.Remove != "" {
		el.Find(s.Remove).Remove()
	}

	if s.Case != nil {
		for pattern, value := range s.Case {
			if el.Is(pattern) || el.Has(pattern).Length() >= 1 {
				return s.FilterText(value)
			}
		}
		return "", errors.New("none of the cases match")
	}
	output := strings.TrimSpace(el.Text())
	output = spaceRx.ReplaceAllString(output, " ")
	if s.Attribute != "" {
		val, exists := el.Attr(s.Attribute)
		if !exists {
			return "", fmt.Errorf("requested attribute %q doesn't exist", s.Attribute)
		}
		output = val
	}

	return s.FilterText(output)
}

// Filter the value through a list of filters
func (s *SelectorBlock) FilterText(val string) (string, error) {
	prevFilterFailed := false
	var prevFilter filterBlock
	for _, f := range s.Filters {
		var shouldFilter bool
		switch f.Name {
		case "dateparseAlt":
			// This is ran only if there has been a failure before.
			shouldFilter = prevFilterFailed && prevFilter.Name == "dateparse"
		default:
			shouldFilter = true
		}
		if !shouldFilter {
			continue
		}

		var err error
		newVal, err := filterService.Filter(f.Name, f.Args, val)
		if err != nil {
			if f.Name != "dateparse" {
				logrus.
					WithFields(logrus.Fields{"selector": s.Selector}).
					Warningf("Filter %s(%s) failed on value `%v`. %s\n", f.Name, f.Args, val, err)
			}
			prevFilterFailed = true
			prevFilter = f
			continue
			// return "", err
		}
		// If we've got a template
		if strings.Contains(newVal, "{{") {
			filterContext := struct {
				Config map[string]string
			}{
				s.FilterConfig,
			}
			newVal, err = templates.ApplyTemplate("filter_template", newVal, filterContext)
			if err != nil {
				// We revert back..
				newVal = val
			}
		}

		val = newVal
		prevFilterFailed = false
	}
	val = strings.TrimSpace(val)
	val = spaceRx.ReplaceAllString(val, " ")
	return val, nil
}

var spaceRx = regexp.MustCompile(`\s+`)

func (s *SelectorBlock) IsEmpty() bool {
	return s.Selector == "" && s.TextVal == ""
}

func (s *SelectorBlock) String() string {
	switch {
	case s.Selector != "":
		return fmt.Sprintf("Selector(%s)", s.Selector)
	case s.TextVal != "":
		return fmt.Sprintf("Text(%s)", s.TextVal)
	default:
		return "Empty"
	}
}
