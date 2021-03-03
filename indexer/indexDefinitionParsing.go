package indexer

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/sp0x/torrentd/indexer/source"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/torznab"
)

var defaultRateLimit = 500

const (
	searchEntity = "search"
)

type Definition struct {
	Site         string            `yaml:"site"`
	Version      string            `yaml:"version"`
	Scheme       string            `yaml:"scheme"`
	Settings     []settingsField   `yaml:"settings"`
	Name         string            `yaml:"name"`
	Description  string            `yaml:"description"`
	Language     string            `yaml:"language"`
	Links        stringorslice     `yaml:"links"`
	Capabilities capabilitiesBlock `yaml:"caps"`
	Login        loginBlock        `yaml:"login"`
	Ratio        ratioBlock        `yaml:"ratio"`
	Search       searchBlock       `yaml:"search"`
	stats        DefinitionStats   `yaml:"-"`
	Encoding     string            `yaml:"encoding"`
	// Entities that the index contains
	Entities []entityBlock `yaml:"entities"`
	// The ms to wait between each request.
	RateLimit int `yaml:"ratelimit"`
}

type DefinitionStats struct {
	Size    int64
	ModTime time.Time
	Hash    string
	Source  string
}

func (id *Definition) Stats() DefinitionStats {
	return id.stats
}

func (id *Definition) getNewResultItem() search.ResultItemBase {
	if id.Scheme == "" {
		return &search.ScrapeResultItem{
			ModelData: make(map[string]interface{}),
		}
	}
	if id.Scheme == schemeTorrent {
		return &search.TorrentResultItem{
			ScrapeResultItem: search.ScrapeResultItem{
				ModelData: make(map[string]interface{}),
			},
		}
	}
	return &search.ScrapeResultItem{
		ModelData: make(map[string]interface{}),
	}
}

// getSearchEntity gets the entity that's returned from a search.
func (id *Definition) getSearchEntity() *entityBlock {
	entity := &entityBlock{}
	entity.Name = searchEntity
	key := id.Search.Key
	localizedKey := make([]string, len(key))
	t := reflect.ValueOf(search.ScrapeResultItem{})
	for ix, k := range key {
		field := t.FieldByName(k)
		if !field.IsValid() {
			k = fmt.Sprintf("ExtraFields.%s", k)
		}
		localizedKey[ix] = k
	}
	entity.IndexKey = localizedKey
	return entity
}

type settingsField struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Label string `yaml:"label"`
}

// ParseDefinitionFile loads an Indexer's definition from a file
func ParseDefinitionFile(f *os.File) (*Definition, error) {
	b, err := ioutil.ReadFile(f.Name())
	if err != nil {
		return nil, err
	}

	def, err := ParseDefinition(b)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	def.stats.ModTime = fi.ModTime()
	return def, err
}

func ParseDefinition(src []byte) (*Definition, error) {
	def := Definition{
		Language:     "en-us",
		Encoding:     "utf-8",
		Capabilities: capabilitiesBlock{},
		Login: loginBlock{
			FormSelector: "form",
			Inputs:       inputsBlock{},
		},
		Search: searchBlock{},
	}

	if err := yaml.Unmarshal(src, &def); err != nil {
		return nil, err
	}

	if len(def.Settings) == 0 {
		def.Settings = defaultSettingsFields()
	}

	def.stats = DefinitionStats{
		Size:    int64(len(src)),
		ModTime: time.Now(),
		Hash:    fmt.Sprintf("%x", sha1.Sum(src)),
	}
	if def.RateLimit == 0 {
		def.RateLimit = defaultRateLimit
	}
	return &def, nil
}

func defaultSettingsFields() []settingsField {
	return []settingsField{
		{Name: "username", Label: "Username", Type: "text"},
		{Name: "password", Label: "Password", Type: "password"},
	}
}

type inputsBlock map[string]string

type errorBlockOrSlice []errorBlock

// UnmarshalYAML implements the Unmarshaller interface.
func (e *errorBlockOrSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var blockType errorBlock
	if err := unmarshal(&blockType); err == nil {
		*e = errorBlockOrSlice{blockType}
		return nil
	}

	var sliceType []errorBlock
	if err := unmarshal(&sliceType); err == nil {
		*e = errorBlockOrSlice(sliceType)
		return nil
	}

	return errors.New("failed to unmarshal errorBlockOrSlice")
}

type errorBlock struct {
	Path     string               `yaml:"path"`
	Selector string               `yaml:"selector"`
	Message  source.SelectorBlock `yaml:"message"`
}

func (e *errorBlock) matchPage(res *source.HTMLFetchResult) bool {
	if e.Path != "" {
		return e.Path == res.HTTPResult.URL().Path
	} else if e.Selector != "" {
		return res.Find(e.Selector).Length() > 0
	}
	return false
}

func scrapeItemFromHTML(res *source.HTMLFetchResult) source.RawScrapeItem {
	return &source.DomScrapeItem{Selection: res.DOM.First()}
}

func (e *errorBlock) errorText(from *source.HTMLFetchResult) (string, error) {
	if !e.Message.IsEmpty() {
		matchError, err := e.Message.Match(scrapeItemFromHTML(from))
		return matchError.(string), err
	} else if e.Selector != "" {
		errs := from.Find(e.Selector)
		if errs.Length() < 1 {
			return "error with unmatching selector", nil
		}
		return errs.Get(0).Text(), nil
	}
	return "", errors.New("error declaration must have either Message block or Selection")
}

type pageTestBlock struct {
	Path     string `yaml:"path"`
	Selector string `yaml:"selector"`
}

func (t *pageTestBlock) IsEmpty() bool {
	return t.Path == "" && t.Selector == ""
}

const (
	loginMethodPost   = "post"
	loginMethodForm   = "form"
	loginMethodCookie = "cookie"
	schemeTorrent     = "torrent"
)

type loginBlock struct {
	Path         string            `yaml:"path"`
	FormSelector string            `yaml:"form"`
	Method       string            `yaml:"method"`
	Inputs       inputsBlock       `yaml:"inputs,omitempty"`
	Error        errorBlockOrSlice `yaml:"error,omitempty"`
	Test         pageTestBlock     `yaml:"test,omitempty"`
	Init         initBlock         `yaml:"init,omitempty"`
}

func (l *loginBlock) IsEmpty() bool {
	return l.Path == "" && l.Method == ""
}

func (l *loginBlock) hasError(res source.FetchResult) error {
	webResult, ok := res.(*source.HTMLFetchResult)
	if !ok {
		return errors.New("login works only with web pages")
	}
	for _, e := range l.Error {
		if e.matchPage(webResult) {
			msg, err := e.errorText(webResult)
			if err != nil {
				return err
			}
			return errors.New(strings.TrimSpace(msg))
		}
	}

	return nil
}

type initBlock struct {
	Path string `yaml:"path"`
}

func (init *initBlock) IsEmpty() bool {
	return init.Path == ""
}

type fieldBlock struct {
	Field string
	Block source.SelectorBlock
}

type fieldsListBlock []fieldBlock

func (f *fieldsListBlock) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Unmarshal as a MapSlice to preserve order of fields
	var fields yaml.MapSlice
	if err := unmarshal(&fields); err != nil {
		return errors.New("failed to unmarshal fieldsListBlock")
	}

	// FIXME: there has got to be a better way to do this
	for _, item := range fields {
		b, err := yaml.Marshal(item.Value)
		if err != nil {
			return err
		}
		var sb source.SelectorBlock
		if err = yaml.Unmarshal(b, &sb); err != nil {
			return err
		}
		if sb.FilterConfig == nil {
			sb.FilterConfig = defaultFilterConfig()
		}
		*f = append(*f, fieldBlock{
			Field: item.Key.(string),
			Block: sb,
		})
	}

	return nil
}

type rowsBlock struct {
	source.SelectorBlock
	After       int                  `yaml:"after"`
	Remove      string               `yaml:"remove"`
	DateHeaders source.SelectorBlock `yaml:"dateheaders"`
}

func (r *rowsBlock) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sb source.SelectorBlock
	if err := unmarshal(&sb); err != nil {
		return errors.New("failed to unmarshal rowsBlock")
	}

	var rb struct {
		After       int                  `yaml:"after"`
		Remove      string               `yaml:"remove"`
		DateHeaders source.SelectorBlock `yaml:"dateheaders"`
	}
	if err := unmarshal(&rb); err != nil {
		return errors.New("failed to unmarshal rowsBlock")
	}

	r.After = rb.After
	r.DateHeaders = rb.DateHeaders
	r.SelectorBlock = sb
	r.Remove = rb.Remove
	return nil
}

type capabilitiesBlock struct {
	CategoryMap categoryMap
	SearchModes []search.Mode
}

// UnmarshalYAML implements the Unmarshaller interface.
func (c *capabilitiesBlock) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var intermediate struct {
		Categories map[string]string        `yaml:"indexCategories"`
		Modes      map[string]stringorslice `yaml:"modes"`
	}

	if err := unmarshal(&intermediate); err == nil {
		c.CategoryMap = categoryMap{}
		// Map the found indexCategories using our own Categories `torznab.AllCategories`.
		allCats := categories.AllCategories
		for id, catName := range intermediate.Categories {
			matchedCat := false
			for key, cat := range allCats {
				cat = allCats[key]
				if cat.Name == catName {
					c.CategoryMap[id] = cat
					matchedCat = true
					break
				}
			}
			if !matchedCat {
				logrus.
					WithFields(logrus.Fields{"name": catName, "id": id}).
					Warn("Unknown category")
				continue
				// return fmt.Errorf("Unknown category %q", catName)
			}
		}

		c.SearchModes = []search.Mode{}

		for key, supported := range intermediate.Modes {
			c.SearchModes = append(c.SearchModes, search.Mode{Key: key, Available: true, SupportedParams: supported})
		}

		return nil
	}

	return errors.New("failed to unmarshal capabilities block")
}

// ToTorznab converts a capabilities def. block to torznab capabilities object.
func (c *capabilitiesBlock) ToTorznab() torznab.Capabilities {
	caps := torznab.Capabilities{
		Categories:  c.CategoryMap.Categories(),
		SearchModes: []search.Mode{},
	}

	// All indexesCollection support search
	caps.SearchModes = append(caps.SearchModes, search.Mode{
		Key:             "search",
		Available:       true,
		SupportedParams: []string{"q"},
	})

	// Some support TV
	if caps.HasTVShows() {
		caps.SearchModes = append(caps.SearchModes, search.Mode{
			Key:             "tv-search",
			Available:       true,
			SupportedParams: []string{"q", "season", "ep"},
		})
	}

	// Some support Movies
	if caps.HasMovies() {
		caps.SearchModes = append(caps.SearchModes, search.Mode{
			Key:             "movie-search",
			Available:       true,
			SupportedParams: []string{"q"},
		})
	}

	return caps
}

type stringorslice []string

func (s *stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringType string
	if err := unmarshal(&stringType); err == nil {
		*s = stringorslice{stringType}
		return nil
	}

	var sliceType []string
	if err := unmarshal(&sliceType); err == nil {
		*s = stringorslice(sliceType)
		return nil
	}

	return errors.New("failed to unmarshal stringorslice")
}

type ratioBlock struct {
	source.SelectorBlock
	Path string `yaml:"path"`
}

func (r *ratioBlock) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sb source.SelectorBlock
	if err := unmarshal(&sb); err != nil {
		return errors.New("failed to unmarshal ratioBlock")
	}

	var rb struct {
		Path string `yaml:"path"`
	}
	if err := unmarshal(&rb); err != nil {
		return errors.New("failed to unmarshal ratioBlock")
	}

	r.SelectorBlock = sb
	r.Path = rb.Path
	return nil
}

func defaultFilterConfig() map[string]string {
	return map[string]string{
		"striprussian": "true",
	}
}
