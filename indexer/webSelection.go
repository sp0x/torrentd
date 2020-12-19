package indexer

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type filterBlock struct {
	Name string      `yaml:"name"`
	Args interface{} `yaml:"args"`
}

type selectorBlock struct {
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
	//If we'll use all the values
	All bool `yaml:"all"`
}

func (s *selectorBlock) IsMatching(selection *goquery.Selection) bool {
	return !s.IsEmpty() && (selection.Find(s.Selector).Length() > 0 || s.TextVal != "")
}

//Match using the selector to get the text of that element
func (s *selectorBlock) Match(from RawScrapeItem) (interface{}, error) {
	if s.TextVal != "" {
		return s.TextVal, nil
	}
	if s.Selector != "" {
		result := from.Find(s)
		if result.Length() == 0 {
			return "", fmt.Errorf("Failed to match selector %q", s.Selector)
		}
		if s.All {
			return s.Texts(result)
		} else {
			return s.Text(result)
		}
	}
	if s.Pattern != "" {
		return s.Pattern, nil
	}
	return s.Text(from)
}

func (s *selectorBlock) MatchRawText(from *goquery.Selection) (string, error) {
	if s.TextVal != "" {
		return s.TextVal, nil
	}
	if s.Selector != "" {
		result := from.Find(s.Selector)
		if result.Length() == 0 {
			return "", fmt.Errorf("Failed to match selector %q", s.Selector)
		}
		return s.TextRaw(result)
	}
	if s.Pattern != "" {
		return s.Pattern, nil
	}
	return s.TextRaw(from)
}

func (s *selectorBlock) TextRaw(el *goquery.Selection) (string, error) {
	if s.TextVal != "" {
		return s.TextVal, nil
	}
	if s.Remove != "" {
		el.Find(s.Remove).Remove()
	}
	if s.Case != nil {
		//filterLogger.WithFields(logrus.Fields{"case": s.Case}).
		//	Debugf("Applying case to selection")
		for pattern, value := range s.Case {
			if el.Is(pattern) || el.Has(pattern).Length() >= 1 {
				return s.ApplyFilters(value)
			}
		}
		return "", errors.New("None of the cases match")
	}
	output := strings.TrimSpace(el.Text())
	output = spaceRx.ReplaceAllString(output, " ")
	if s.Attribute != "" {
		val, exists := el.Attr(s.Attribute)
		if !exists {
			return "", fmt.Errorf("Requested attribute %q doesn't exist", s.Attribute)
		}
		output = val
	}
	return output, nil
}

func (s *selectorBlock) Texts(element RawScrapeItem) ([]string, error) {
	if s.TextVal != "" {
		filterResult, err := s.ApplyFilters(s.TextVal)
		if err != nil {
			return nil, err
		}
		return []string{filterResult}, nil
	}
	if s.Remove != "" {
		element.Find(s.Remove).Remove()
	}
	if s.Case != nil {
		filterLogger.WithFields(logrus.Fields{"case": s.Case}).
			Debugf("Applying case to selection")
		for pattern, value := range s.Case {
			if element.Is(pattern) || element.Has(pattern).Length() >= 1 {
				value, _ := s.ApplyFilters(value)
				return []string{value}, nil
			}
		}
		return []string{}, errors.New("none of the cases match")
	}
	matches := element.Map(func(i int, selection *goquery.Selection) string {
		output := strings.TrimSpace(selection.Text())
		output = spaceRx.ReplaceAllString(output, " ")
		if s.Attribute != "" {
			val, exists := selection.Attr(s.Attribute)
			if !exists {
				return ""
			}
			output = val
		}
		filteredResult, err := s.ApplyFilters(output)
		if err != nil {
			return ""
		}
		return filteredResult
	})
	return matches, nil
}

//Text extracts text from the selection, applying all filters
func (s *selectorBlock) Text(el *goquery.Selection) (string, error) {
	if s.TextVal != "" {
		return s.ApplyFilters(s.TextVal)
	}

	if s.Remove != "" {
		el.Find(s.Remove).Remove()
	}

	if s.Case != nil {
		filterLogger.WithFields(logrus.Fields{"case": s.Case}).
			Debugf("Applying case to selection")
		for pattern, value := range s.Case {
			if el.Is(pattern) || el.Has(pattern).Length() >= 1 {
				return s.ApplyFilters(value)
			}
		}
		return "", errors.New("None of the cases match")
	}
	output := strings.TrimSpace(el.Text())
	output = spaceRx.ReplaceAllString(output, " ")
	if s.Attribute != "" {
		val, exists := el.Attr(s.Attribute)
		if !exists {
			return "", fmt.Errorf("Requested attribute %q doesn't exist", s.Attribute)
		}
		output = val
	}

	return s.ApplyFilters(output)
}

//Filter the value through a list of filters
func (s *selectorBlock) ApplyFilters(val string) (string, error) {
	prevFilterFailed := false
	var prevFilter filterBlock
	for _, f := range s.Filters {
		var shouldFilter bool
		switch f.Name {
		case "dateparseAlt":
			//This is ran only if there has been a failure before.
			shouldFilter = prevFilterFailed && prevFilter.Name == "dateparse"
		default:
			shouldFilter = true
		}
		if !shouldFilter {
			continue
		}

		filterLogger.WithFields(logrus.Fields{"args": f.Args, "before": val}).
			Debugf("Applying filter %s", f.Name)
		var err error
		newVal, err := invokeFilter(f.Name, f.Args, val)
		if err != nil {
			if f.Name != "dateparse" {
				logrus.
					WithFields(logrus.Fields{"selector": s.Selector}).
					Warningf("Filter %s(%s) failed on value `%v`. %s\n", f.Name, f.Args, val, err)
			}
			prevFilterFailed = true
			prevFilter = f
			continue
			//return "", err
		}
		//If we've got a template
		if strings.Contains(newVal, "{{") {
			filterContext := struct {
				Config map[string]string
			}{
				s.FilterConfig,
			}
			newVal, err = applyTemplate("filter_template", newVal, filterContext)
			if err != nil {
				//We revert back..
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

func (s *selectorBlock) IsEmpty() bool {
	return s.Selector == "" && s.TextVal == ""
}

func (s *selectorBlock) String() string {
	switch {
	case s.Selector != "":
		return fmt.Sprintf("Selector(%s)", s.Selector)
	case s.TextVal != "":
		return fmt.Sprintf("Text(%s)", s.TextVal)
	default:
		return "Empty"
	}
}
