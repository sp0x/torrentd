package indexer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/source"
	"github.com/sp0x/torrentd/indexer/source/series"
	"github.com/sp0x/torrentd/indexer/status"
	"github.com/sp0x/torrentd/indexer/utils"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/torznab"
)

var (
	_        Indexer = &Runner{}
	errorTTL         = 2 * 24 * time.Hour
)

const (
	indexVerificationSpan = time.Minute * 60 * 24
	errorValue            = "error"
)

type RunnerOpts struct {
	Config       config.Config
	CachePages   bool
	Transport    http.RoundTripper
	UserSessions int
}

// Runner works index definitions in order to extract data.
type Runner struct {
	definition          *Definition
	options             *RunnerOpts
	logger              log.FieldLogger
	connectivityTester  cache.ConnectivityTester
	failingSearchFields map[string]fieldBlock
	lastVerified        time.Time
	contentFetcher      source.ContentFetcher
	context             context.Context
	errors              cache.LRUCache
	sessions            *BrowsingSessionMultiplexer
	statusReporter      *StatusReporter
	urlResolver         IURLResolver
}

type scrapeContext struct {
	query           *search.Query
	indexCategories []string
}

func (r *Runner) GetDefinition() *Definition {
	return r.definition
}

func (r *Runner) Site() string {
	if r.definition == nil {
		return ""
	}
	return r.definition.Name
}

func (r *Runner) MaxSearchPages() uint {
	p := uint(r.definition.Search.MaxPages)
	if r.SearchIsSinglePaged() {
		return 1
	}
	return p
}

func (r *Runner) SearchIsSinglePaged() bool {
	return r.definition.Search.IsSinglePage()
}

func getConfiguredIndexLoader(conf config.Config) DefinitionLoader {
	loader := conf.Get("indexLoader")
	if loader == nil {
		return GetIndexDefinitionLoader()
	}
	return loader.(DefinitionLoader)
}

func setConfiguredIndexLoader(loader DefinitionLoader, conf config.Config) {
	conf.Set("indexLoader", loader)
}

// NewIndexRunnerByNameOrSelector creates a new Indexer or aggregate Indexer with the given configuration.
func NewIndexRunnerByNameOrSelector(indexerName string, config config.Config) (IndexCollection, error) {
	def, err := getConfiguredIndexLoader(config).Load(indexerName)
	if err != nil {
		log.WithError(err).Warnf("Failed to load definition for %q. %v", indexerName, err)
		return nil, err
	}

	index := NewRunner(def, runnerOptsFromConfig(config))
	return IndexCollection{index}, nil
}

func runnerOptsFromConfig(config config.Config) *RunnerOpts {
	userSessions := config.GetInt("users")
	if userSessions <= 0 {
		userSessions = 1
	}
	workers := config.GetInt("workers")
	if workers <= 0 {
		workers = 1
	}

	opts := &RunnerOpts{
		Config:       config,
		UserSessions: userSessions,
	}
	return opts
}

// NewRunner Start a runner for a given indexer.
func NewRunner(def *Definition, opts *RunnerOpts) *Runner {
	logger := log.New().WithFields(log.Fields{"site": def.Site})
	logger.Level = log.GetLevel()
	// Use an optimistic cache instead.
	errorCache, _ := cache.NewTTL(10, errorTTL)
	indexCtx := context.Background()

	runner := &Runner{
		options:             opts,
		definition:          def,
		logger:              logger,
		failingSearchFields: make(map[string]fieldBlock),
		context:             indexCtx,
		errors:              errorCache,
		statusReporter:      &StatusReporter{context: indexCtx, indexDefinition: def, errors: errorCache},
	}
	runner.contentFetcher = createContentFetcher(runner)
	connectivity, _ := cache.NewConnectivityCache(runner.contentFetcher)
	runner.urlResolver = newURLResolverForIndex(def, opts.Config, connectivity)
	runner.connectivityTester = connectivity
	runner.contentFetcher.SetErrorHandler(func(options *source.RequestOptions) {
		connectivity.Invalidate(options.URL.String())
	})

	if sessionsMx, err := NewSessionMultiplexer(runner, opts.UserSessions); err != nil {
		fmt.Printf("Couldn't create index sessions: %v", err)
		os.Exit(1)
	} else {
		runner.sessions = sessionsMx
	}
	return runner
}

func (r *Runner) GetStorage() storage.ItemStorage {
	itemStorage := getIndexDatabase(r.definition.Name, r.definition.getSearchEntity(), r.options.Config)
	return itemStorage
}

func getMultiIndexDatabase(indexes IndexCollection, conf config.Config) storage.ItemStorage {
	return getIndexDatabase(indexes.Name(), nil, conf)
}

func getIndexDatabase(name string, searchEntityBlock *entityBlock, conf config.Config) storage.ItemStorage {
	storageType := conf.GetString("storage")
	if storageType == "" {
		panic("no database type configured")
	}
	var itemStorage storage.ItemStorage
	dbEndpoint := conf.GetString("storageendpoint")
	if dbEndpoint == "" {
		panic("no database endpoint configured")
	}
	if searchEntityBlock != nil {
		itemStorage = storage.NewBuilder().
			WithNamespace(name).
			WithEndpoint(dbEndpoint).
			WithPK(searchEntityBlock.GetKey()).
			WithBacking(storageType).
			WithRecord(&search.ScrapeResultItem{}).
			Build()
	} else {
		itemStorage = storage.NewBuilder().
			WithNamespace(name).
			WithEndpoint(dbEndpoint).
			WithBacking(storageType).
			WithRecord(&search.ScrapeResultItem{}).
			Build()
	}

	return itemStorage
}

// Capabilities gets the torznab formatted capabilities of this Indexer.
func (r *Runner) Capabilities() torznab.Capabilities {
	caps := r.definition.Capabilities.ToTorznab()

	for idx, mode := range caps.SearchModes {
		switch mode.Key {
		case "search":
			caps.SearchModes[idx].SupportedParams = append(
				caps.SearchModes[idx].SupportedParams,
				"imdbid", "tvdbid", "tvmazeid")

		case "movie-search":
			caps.SearchModes[idx].SupportedParams = append(
				caps.SearchModes[idx].SupportedParams,
				"imdbid")

		case "tv-search":
			caps.SearchModes[idx].SupportedParams = append(
				caps.SearchModes[idx].SupportedParams,
				"tvdbid", "tvmazeid", "rid")
		}
	}

	return caps
}

// GetEncoding returns the encoding that's set to be used in this index.
// This can be changed in the index's definition.
func (r *Runner) GetEncoding() string {
	return r.definition.Encoding
}

// HealthCheck checks if the index can be searched.
// Health checks for each index have a duration of 1 day.
func (r *Runner) HealthCheck() error {
	verifiedSpan := time.Since(r.lastVerified)
	if verifiedSpan < indexVerificationSpan {
		return nil
	}
	results, err := r.Search(search.NewQuery(), nil)
	if err != nil {
		return err
	}
	if results == nil {
		return fmt.Errorf("search result was null")
	}

	if len(results) == 0 {
		return fmt.Errorf("failed health check. no items returned")
	}

	r.lastVerified = time.Now()
	return nil
}

//func getUniqueIndex(item search.ResultItemBase) *indexing.Key {
//	if item == nil {
//		return nil
//	}
//	key := indexing.NewKey()
//	scrapeItem := item.AsScrapeItem()
//	// Local id would be a good bet.
//	if len(scrapeItem.LocalID) > 0 {
//		key.Add("LocalID")
//	}
//	return key
//}

// Search for a given torrent
func (r *Runner) Search(query *search.Query, srch *workerJob) ([]search.ResultItemBase, error) {
	var err error
	errType := status.LoginError
	// Collect errors on exit
	defer func() { r.noteError(errType, err) }()

	session, err := r.sessions.acquire()
	if err != nil {
		return nil, err
	}

	categories := GetLocalCategoriesMatchingQuery(query, &r.definition.Capabilities)
	reqOpts, err := r.createRequest(query, categories, srch, session)
	if err != nil {
		errType = status.TargetError
		r.logger.WithError(err).Warn(err)
		return nil, err
	}
	startedOn := time.Now()
	// Search the content
	fetchResult, err := r.contentFetcher.Fetch(reqOpts)
	if err != nil {
		errType = status.ContentError
		return nil, err
	}
	scrapeItems, err := r.extractScrapeItems(fetchResult, srch)
	if scrapeItems == nil || err != nil {
		errType = status.ContentError
		return nil, fmt.Errorf("result items could not be enumerated.%v", err)
	}
	rowContext := &scrapeContext{
		query,
		categories,
	}

	results := r.processScrapedItems(scrapeItems, rowContext)

	r.logger.
		WithFields(log.Fields{
			"Indexes": r.definition.Site,
			"search":  srch.String(),
			"q":       query.Keywords(),
			"time":    time.Since(startedOn),
		}).
		Infof("Query returned %d results", len(results))

	status.PublishSchemeStatus(r.context, generateSchemeOkStatus(r.definition, results))
	return results, nil
}

// Goes through the scraped items and converts them to the defined data structure
func (r *Runner) processScrapedItems(scrapeItems source.RawScrapeItems, rowContext *scrapeContext) []search.ResultItemBase {
	var results []search.ResultItemBase
	for i := 0; i < scrapeItems.Length(); i++ {
		if rowContext.query.HasEnoughResults(len(results)) {
			break
		}
		// Search the result from the row
		item, err := r.extractItem(i+1, scrapeItems.Get(i), rowContext)
		if err != nil {
			r.logger.Errorf("Couldn't extract item: %v", err)
			continue
		}

		results = append(results, item)
	}
	return results
}

func (r *Runner) resolveItemCategory(query *search.Query, localCats []string, item search.ResultItemBase) bool {
	if !itemMatchesScheme("torrent", item) {
		return false
	}
	torrentItem := item.(*search.TorrentResultItem)
	if len(localCats) > 0 {
		// The category doesn't match even 1 of the indexCategories in the query.
		if !r.itemMatchesLocalCategories(localCats, torrentItem) {
			r.logger.
				WithFields(log.Fields{
					"category":   torrentItem.LocalCategoryName,
					"categoryId": torrentItem.LocalCategoryID,
				}).
				Debugf("Skipping result because it's not contained in our needed indexCategories.")
			return false
		}
	}
	// Try to map the category from the Indexes to the global indexCategories
	r.populateCategory(item)
	return !series.IsSeriesAndNotMatching(query, torrentItem)
}

func (r *Runner) clearDom(dom *goquery.Selection) error {
	searchDef := r.definition.Search
	if after := searchDef.Rows.After; after > 0 {
		rows := dom.Find(searchDef.Rows.Selector)
		for i := 0; i < rows.Length(); i += 1 + after {
			rows.Eq(i).AppendSelection(rows.Slice(i+1, i+1+after).Find("td"))
			rows.Slice(i+1, i+1+after).Remove()
		}
	}
	// apply Remove if it exists
	if remove := searchDef.Rows.Remove; remove != "" {
		matching := dom.Find(searchDef.Rows.Selector).Filter(remove)
		r.logger.
			WithFields(log.Fields{"selector": remove}).
			Debugf("Applying remove to %d rows", matching.Length())
		matching.Remove()
	}
	if searchDef.Rows.Selector == "" {
		return errors.New("no result item selector is given")
	}
	return nil
}

func (r *Runner) createRequest(query *search.Query, lCategories []string, srch *workerJob, session *BrowsingSession) (*source.RequestOptions, error) {
	// Exposed fields to add:
	templateData := r.getSearchTemplateData(query, srch, lCategories)
	// ApplyTo our context to the search path
	initialSrcURL, err := templateData.ApplyTo("search_path", r.definition.Search.Path)
	if err != nil {
		return nil, err
	}
	// Resolve the search url
	searchURL, err := r.urlResolver.Resolve(initialSrcURL)
	if err != nil {
		return nil, err
	}
	// Search our Indexes url values
	vals, err := getURLValuesForSearch(&r.definition.Search, templateData)
	if err != nil {
		return nil, err
	}
	req := &source.RequestOptions{
		URL:    searchURL,
		Values: vals,
		Method: r.definition.Search.Method,
	}
	if session != nil {
		session.ApplyToRequest(req)
	}
	return req, nil
}

func getURLValuesForSearch(searchDef *searchBlock, templateData *SearchTemplateData) (url.Values, error) {
	// Parse the values that will be used in the url for the search
	urlValues := url.Values{}

	for inputFieldName, inputValueFromScheme := range searchDef.Inputs {
		if templateData.HasQueryField(inputFieldName) {
			inputValueFromScheme, _ = templateData.GetSearchFieldValue(inputFieldName)
		}
		resolveInputValue, err := templateData.ApplyTo(inputFieldName, inputValueFromScheme)
		if err != nil {
			return nil, err
		}
		switch inputFieldName {
		case "$raw":
			err := evalRawSearchInputs(resolveInputValue, urlValues)
			if err != nil {
				return urlValues, err
			}
		default:
			urlValues.Add(inputFieldName, resolveInputValue)
		}
	}
	return urlValues, nil
}

func evalRawSearchInputs(resolvedInputValue string, vals url.Values) error {
	parsedVals, err := url.ParseQuery(resolvedInputValue)
	if err != nil {
		return fmt.Errorf("error parsing $raw input: %s", err.Error())
	}

	for k, values := range parsedVals {
		for _, val := range values {
			vals.Add(k, val)
		}
	}
	return nil
}

// Search the default run context
func (r *Runner) getSearchTemplateData(query *search.Query, srch *workerJob, lCategories []string) *SearchTemplateData {
	// startIndex := int(query.Page) * r.definition.Search.PageSize
	// TODO: srch.SetStartIndex(r, startIndex)
	data := newSearchTemplateData(query, srch, lCategories)
	return data
}

func (r *Runner) hasDateHeader() bool {
	return !r.definition.Search.Rows.DateHeaders.IsEmpty()
}

func (r *Runner) extractDateHeader(selection source.RawScrapeItem) (time.Time, error) {
	dateHeaders := r.definition.Search.Rows.DateHeaders

	r.logger.
		WithFields(log.Fields{"selector": dateHeaders.String()}).
		Debugf("Searching for date header")

	prev := selection.PrevAllFiltered(dateHeaders.Selector).First()
	if prev.Length() == 0 {
		return time.Time{}, fmt.Errorf("no date header row found")
	}

	dv, _ := dateHeaders.Text(prev.First())
	return utils.ParseFuzzyTime(dv, time.Now(), true)
}

func (r *Runner) Ratio() (string, error) {
	if r.definition.Ratio.TextVal != "" {
		return r.definition.Ratio.TextVal, nil
	}

	if r.definition.Ratio.Path == "" {
		return "unknown", nil
	}

	if _, err := r.sessions.acquire(); err != nil {
		return errorValue, err
	}

	ratioURL, err := r.urlResolver.Resolve(r.definition.Ratio.Path)
	if err != nil {
		return errorValue, err
	}

	resultData, err := r.contentFetcher.Fetch(source.NewRequestOptions(ratioURL))
	if err != nil {
		r.logger.WithError(err).Warn("Failed to open page")
		return errorValue, nil
	}

	var ratio interface{}
	switch value := resultData.(type) {
	case *source.HTMLFetchResult:
		ratio, err := r.definition.Ratio.Match(&source.DomScrapeItem{Selection: value.DOM.First()})
		if err != nil {
			return ratio.(string), err
		}
	default:
		return "", errors.New("response was not html")
	}

	return strings.Trim(ratio.(string), "- "), nil
}

func (r *Runner) getIndexer() *search.ResultIndexer {
	return &search.ResultIndexer{
		ID:   "",
		Name: r.definition.Name,
	}
}

func (r *Runner) Errors() []string {
	return r.statusReporter.GetErrors()
}

func (r *Runner) noteError(errorType string, err error) {
	r.statusReporter.Error(NewError(errorType, err))
}

// region Status messages

func generateSchemeOkStatus(definition *Definition, searchResults []search.ResultItemBase) *status.ScrapeSchemeMessage {
	statusCode := "ok"
	resultsFound := 0
	if len(searchResults) > 0 {
		statusCode = "ok-data"
		resultsFound = len(searchResults)
	}
	msg := &status.ScrapeSchemeMessage{
		Code:          statusCode,
		Site:          definition.Site,
		SchemeVersion: definition.Version,
		ResultsFound:  resultsFound,
	}
	return msg
}

func generateSchemeErrorStatus(errorCode string, err error, definition *Definition) *status.SchemeErrorMessage {
	msg := &status.SchemeErrorMessage{
		Code:          errorCode,
		Site:          definition.Site,
		SchemeVersion: definition.Version,
		Message:       fmt.Sprintf("couldn't log in: %s", err),
	}
	return msg
}

// endregion
