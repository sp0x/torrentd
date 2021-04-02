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
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/source"
	"github.com/sp0x/torrentd/indexer/source/series"
	"github.com/sp0x/torrentd/indexer/status"
	"github.com/sp0x/torrentd/indexer/utils"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/storage/indexing"
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
	Workers      int
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

type RunContext struct {
	Search *search.Search
}

type scrapeContext struct {
	query           *search.Query
	indexCategories []string
	storage         storage.ItemStorage
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

// NewRunnerByNameOrSelector creates a new Indexer or aggregate Indexer with the given configuration.
func NewRunnerByNameOrSelector(indexerName string, config config.Config) (Indexer, error) {
	def, err := Loader.Load(indexerName)
	if err != nil {
		log.WithError(err).Warnf("Failed to load definition for %q. %v", indexerName, err)
		return nil, err
	}

	indexer := NewRunner(def, runnerOptsFromConfig(config))
	return indexer, nil
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
		Workers:      workers,
	}
	return opts
}

// NewRunner Start a runner for a given indexer.
func NewRunner(def *Definition, opts *RunnerOpts) *Runner {
	logger := log.New()
	logger.Level = log.GetLevel()
	// Use an optimistic cache instead.
	errorCache, _ := cache.NewTTL(10, errorTTL)
	indexCtx := context.Background()

	runner := &Runner{
		options:             opts,
		definition:          def,
		logger:              logger.WithFields(log.Fields{"site": def.Site}),
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
	itemStorage := getIndexStorage(r, r.options.Config)
	return itemStorage
}

func getIndexStorage(indexer Indexer, conf config.Config) storage.ItemStorage {
	definition := indexer.GetDefinition()
	entityType := definition.getSearchEntity()
	storageType := conf.GetString("storage")
	if storageType == "" {
		panic("no storage type configured")
	}
	var itemStorage storage.ItemStorage
	dbPath := conf.GetString("db")
	if entityType != nil {
		// All the results will be stored in a collection with the same name as the index.
		itemStorage = storage.NewBuilder().
			WithNamespace(definition.Name).
			WithEndpoint(dbPath).
			WithPK(entityType.GetKey()).
			WithBacking(storageType).
			WithRecord(&search.ScrapeResultItem{}).
			Build()
	} else {
		// All the results will be stored in a collection with the same name as the index.
		itemStorage = storage.NewBuilder().
			WithNamespace(definition.Name).
			WithEndpoint(dbPath).
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

// getLocalCategoriesMatchingQuery returns a slice of local indexCategories that should be searched
func (r *Runner) getLocalCategoriesMatchingQuery(query *search.Query) []string {
	var localCats []string
	set := make(map[string]struct{})
	if len(query.Categories) > 0 {
		queryCats := categories.AllCategories.Subset(query.Categories...)
		// resolve query indexCategories to the exact local, or the local based on parent cat
		for _, id := range r.definition.Capabilities.CategoryMap.ResolveAll(queryCats.Items()...) {
			// Add only if it doesn't exist
			if _, ok := set[id]; !ok {
				localCats = append(localCats, id)
				set[id] = struct{}{}
			}
		}
	}

	return localCats
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
	searchResult, err := r.Search(search.NewQuery(), nil)
	if err != nil {
		return err
	}
	if searchResult == nil {
		return fmt.Errorf("search result was null")
	}

	resultItems := searchResult.GetResults()
	if len(resultItems) == 0 {
		return fmt.Errorf("failed health check. no items returned")
	}

	r.lastVerified = time.Now()
	return nil
}

func (r *Runner) getUniqueIndex(item search.ResultItemBase) *indexing.Key {
	if item == nil {
		return nil
	}
	key := indexing.NewKey()
	scrapeItem := item.AsScrapeItem()
	// Local id would be a good bet.
	if len(scrapeItem.LocalID) > 0 {
		key.Add("LocalID")
	}
	return key
}

// SearchKeywords for a given torrent
func (r *Runner) Search(query *search.Query, searchInstance search.Instance) (search.Instance, error) {
	if searchInstance == nil {
		return nil, errors.New("search instance is null")
	}
	var err error
	errType := status.LoginError
	// Collect errors on exit
	defer func() { r.noteError(errType, err) }()

	session, err := r.sessions.acquire()
	if err != nil {
		return searchInstance, err
	}

	runCtx := &RunContext{
		Search: searchInstance.(*search.Search),
	}

	localCats := r.getLocalCategoriesMatchingQuery(query)
	reqOpts, err := r.createRequest(query, localCats, runCtx, session)
	if err != nil {
		errType = status.TargetError
		r.logger.WithError(err).Warn(err)
		return nil, err
	}
	timer := time.Now()
	// Get the content
	fetchResult, err := r.contentFetcher.Fetch(reqOpts)
	if err != nil {
		errType = status.ContentError
		return nil, err
	}
	rows, err := r.getRows(fetchResult, runCtx)
	if rows == nil || err != nil {
		errType = status.ContentError
		return nil, fmt.Errorf("result items could not be enumerated.%v", err)
	}
	r.logger.
		WithFields(log.Fields{
			"rows":     rows.Length(),
			"selector": r.definition.Search.Rows,
			"limit":    query.Limit,
			"offset":   query.Offset,
		}).Debugf("Found %d rows", rows.Length())

	rowContext := &scrapeContext{
		query,
		localCats,
		r.GetStorage(),
	}
	defer rowContext.storage.Close()
	results := r.processScrapedItems(rows, rowContext)

	r.logger.
		WithFields(log.Fields{
			"Index":  r.definition.Site,
			"search": runCtx.Search.String(),
			"q":      query.Keywords(),
			"time":   time.Since(timer),
		}).
		Infof("Query returned %d results", len(results))
	runCtx.Search.SetResults(results)
	status.PublishSchemeStatus(r.context, generateSchemeOkStatus(r.definition, runCtx))
	return runCtx.Search, nil
}

// Goes through the scraped items and converts them to the defined data structure
func (r *Runner) processScrapedItems(rows source.RawScrapeItems, rowContext *scrapeContext) []search.ResultItemBase {
	var results []search.ResultItemBase
	for i := 0; i < rows.Length(); i++ {
		if rowContext.query.HasEnoughResults(len(results)) {
			break
		}
		// Get the result from the row
		item, err := r.extractItem(i+1, rows.Get(i), rowContext)
		if err != nil {
			continue
		}
		_ = rowContext.storage.SetKey(r.getUniqueIndex(item))
		err = rowContext.storage.Add(item)
		if err != nil {
			r.logger.Errorf("Couldn't add item: %s\n", err)
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
	// Try to map the category from the Index to the global indexCategories
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

func (r *Runner) createRequest(query *search.Query, lCategories []string, context *RunContext, session *BrowsingSession) (*source.RequestOptions, error) {
	// Exposed fields to add:
	templateData := r.getSearchTemplateData(query, lCategories, context)
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
	// Get our Index url values
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

func getURLValuesForSearch(srchDef *searchBlock, templateData *SearchTemplateData) (url.Values, error) {
	// Parse the values that will be used in the url for the search
	urlValues := url.Values{}

	for inputFieldName, inputValueFromScheme := range srchDef.Inputs {
		if templateData.HasQueryField(inputFieldName) {
			inputValueFromScheme, _ = templateData.ApplyField(inputFieldName)
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

// Get the default run context
func (r *Runner) getSearchTemplateData(query *search.Query, lCategories []string, ctxt *RunContext) *SearchTemplateData {
	startIndex := int(query.Page) * r.definition.Search.PageSize
	ctxt.Search.SetStartIndex(r, startIndex)
	data := newSearchTemplateData(query, lCategories, ctxt)
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

func generateSchemeOkStatus(definition *Definition, runCtx *RunContext) *status.ScrapeSchemeMessage {
	statusCode := "ok"
	resultsFound := 0
	if runCtx.Search != nil && len(runCtx.Search.Results) > 0 {
		statusCode = "ok-data"
		resultsFound = len(runCtx.Search.Results)
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
