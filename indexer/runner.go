package indexer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/source"
	"github.com/sp0x/torrentd/indexer/source/series"
	"github.com/sp0x/torrentd/indexer/source/web"
	"github.com/sp0x/torrentd/indexer/status"
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
	Config     config.Config
	CachePages bool
	Transport  http.RoundTripper
}

// Runner works index definitions in order to extract data.
type Runner struct {
	definition          *Definition
	browser             browser.Browsable
	cookies             http.CookieJar
	options             RunnerOpts
	logger              logrus.FieldLogger
	browserLock         sync.Mutex
	connectivityTester  cache.ConnectivityTester
	state               *State
	keepSessions        bool
	failingSearchFields map[string]fieldBlock
	lastVerified        time.Time
	contentFetcher      source.ContentFetcher
	context             context.Context
	errors              cache.LRUCache
	session             *BrowsingSession
	statusReporter      StatusReporter
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

func (r *Runner) ProcessRequest(req *http.Request) (*http.Response, error) {
	st := r.browser.State()
	st.Request = req
	err := r.browser.Reload()
	return st.Response, err
}

type RunContext struct {
	Search *search.Search
	// SearchKeywords *search.Instance
}

// NewRunner Start a runner for a given indexer.
func NewRunner(def *Definition, opts RunnerOpts) *Runner {
	logger := logrus.New()
	logger.Level = logrus.GetLevel()
	// connCache, _ := cache.NewConnectivityCache()
	// Use an optimistic cache instead.
	connCache, _ := cache.NewOptimisticConnectivityCache()
	errorCache, _ := cache.NewTTL(10, errorTTL)
	indexCtx := context.Background()
	runner := &Runner{
		options:             opts,
		definition:          def,
		logger:              logger.WithFields(logrus.Fields{"site": def.Site}),
		connectivityTester:  connCache,
		state:               defaultIndexerState(),
		keepSessions:        true,
		failingSearchFields: make(map[string]fieldBlock),
		context:             indexCtx,
		errors:              errorCache,
		statusReporter:      StatusReporter{context: indexCtx, indexDefinition: def, errors: errorCache},
	}
	runner.createBrowser()
	if session, err := newIndexSessionFromRunner(runner); err != nil {
		fmt.Printf("Couldn't create index session: %v", err)
		os.Exit(1)
	} else {
		runner.session = session
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
		// runner.Storage = storage.NewKeyedStorageWithBackingType(def.Name, runnerConfig, indexing.NewKey(), storageType)
		itemStorage = storage.NewBuilder().
			WithNamespace(definition.Name).
			WithEndpoint(dbPath).
			WithBacking(storageType).
			WithRecord(&search.ScrapeResultItem{}).
			Build()
	}

	return itemStorage
}

// checks that the runner has the config values it needs
//func (r *Runner) checkHasConfig() error {
//	for _, setting := range r.definition.Settings {
//		_, ok, err := r.options.Config.GetSiteOption(r.definition.IndexName, setting.Name)
//		if err != nil {
//			return fmt.Errorf("Error reading config for %s: %v", setting.Name, err)
//		}
//		if !ok {
//			return fmt.Errorf("No value for %s.%s in config", r.definition.IndexName, setting.Name)
//		}
//	}
//	return nil
//}

// Get a working url for the Indexer
//func (r *Runner) currentURL() (*url.URL, error) {
//	if u := r.browser.Url(); u != nil {
//		return u, nil
//	}
//	configURL, ok, _ := r.options.Config.GetSiteOption(r.definition.Site, "url")
//	if ok && r.testThatUrlWorks(configURL) {
//		return url.Parse(configURL)
//	}
//	for _, u := range r.definition.Links {
//		if u != configURL && r.testThatUrlWorks(u) {
//			return url.Parse(u)
//		}
//	}
//	return nil, errors.New("no working urls found")
//}

// Test if the url returns a 20x response
func (r *Runner) testThatUrlWorks(u string) bool {
	var ok bool
	// Do this like that so it's locked.
	if ok = r.connectivityTester.IsOkAndSet(u, func() bool {
		// The check would be performed only if the connectivity tester doesn't have an entry for that URL
		urlx := u
		r.logger.WithField("url", u).
			Info("Checking connectivity to url")
		err := r.connectivityTester.Test(urlx)
		if err != nil {
			r.logger.WithError(err).Warn("URL check failed")
			return false
		} else if r.browser.StatusCode() != http.StatusOK {
			r.logger.Warn("URL returned non-ok status")
			return false
		}
		return true
	}); ok {
		return true
	}
	return ok
}

// getFullURLInIndex resolve a relative url based on the working Indexer base url
//func (r *Runner) getFullURLInIndex(urlPath string) (string, error) {
//	if strings.HasPrefix(urlPath, "magnet:") {
//		return urlPath, nil
//	}
//	// Get the base url of the Indexer
//	base, err := r.currentURL()
//	if err != nil {
//		return "", err
//	}
//
//	u, err := url.Parse(urlPath)
//	if err != nil {
//		return "", err
//	}
//	// Resolve the url
//	resolved := base.ResolveReference(u)
//	return resolved.String(), nil
//}

func parseCookieString(cookie string) []*http.Cookie {
	h := http.Header{"Cookie": []string{cookie}}
	r := http.Request{Header: h}
	return r.Cookies()
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

		r.logger.
			WithFields(logrus.Fields{"querycats": query.Categories, "local": localCats}).
			Debugf("Resolved torznab cats to local")
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

type scrapeContext struct {
	query           *search.Query
	indexCategories []string
	storage         storage.ItemStorage
}

// SearchKeywords for a given torrent
func (r *Runner) Search(query *search.Query, searchInstance search.Instance) (search.Instance, error) {
	r.createBrowser()
	if !r.keepSessions {
		defer r.releaseBrowser()
	}
	var err error
	// Collect errors on exit
	defer func() { r.noteError(err) }()

	// Login if it's required
	err = r.session.setup()
	if err != nil {
		return searchInstance, err
	}
	// Get the indexCategories for this query based on the Indexer

	// Context about the search
	if searchInstance == nil {
		return nil, errors.New("search instance is null")
		// searchInstance = search.NewSearch(query)
	}
	runCtx := RunContext{
		Search: searchInstance.(*search.Search),
	}

	localCats := r.getLocalCategoriesMatchingQuery(query)
	target, err := r.extractSearchTarget(query, localCats, runCtx)
	if err != nil {
		status.PublishSchemeError(r.context, generateSchemeErrorStatus(status.TargetError, err, r.definition))
		return nil, err
	}
	timer := time.Now()
	// Get the content
	fetchResult, err := r.contentFetcher.Fetch(target)
	if err != nil {
		status.PublishSchemeError(r.context, generateSchemeErrorStatus(status.ContentError, err, r.definition))
		return nil, err
	}
	rows, err := r.getRows(fetchResult, &runCtx)
	if rows == nil {
		return nil, fmt.Errorf("result items could not be enumerated")
	}
	r.logger.
		WithFields(logrus.Fields{
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
		WithFields(logrus.Fields{
			"Indexer": r.definition.Site,
			"search":  runCtx.Search.String(),
			"q":       query.Keywords(),
			"time":    time.Since(timer),
		}).
		Infof("Query returned %d results", len(results))
	runCtx.Search.SetResults(results)
	status.PublishSchemeStatus(r.context, generateSchemeOkStatus(r.definition, runCtx))
	return runCtx.Search, nil
}

// Goes through the scraped items and converts them to the defined data structure
func (r *Runner) processScrapedItems(rows RawScrapeItems, rowContext *scrapeContext) []search.ResultItemBase {
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
				WithFields(logrus.Fields{"category": torrentItem.LocalCategoryName, "categoryId": torrentItem.LocalCategoryID}).
				Debugf("Skipping result because it's not contained in our needed indexCategories.")
			return false
		}
	}
	// Try to map the category from the Indexer to the global indexCategories
	r.populateCategory(item)
	return !series.IsSeriesAndNotMatching(query, torrentItem)
}

func (r *Runner) clearDom(dom *goquery.Selection) error {
	html := r.browser.Body()
	if after := r.definition.Search.Rows.After; after > 0 {
		rows := dom.Find(r.definition.Search.Rows.Selector)
		for i := 0; i < rows.Length(); i += 1 + after {
			rows.Eq(i).AppendSelection(rows.Slice(i+1, i+1+after).Find("td"))
			rows.Slice(i+1, i+1+after).Remove()
		}
	}
	// apply Remove if it exists
	if remove := r.definition.Search.Rows.Remove; remove != "" {
		matching := dom.Find(r.definition.Search.Rows.Selector).Filter(remove)
		r.logger.
			WithFields(logrus.Fields{"selector": remove, "html": html}).
			Debugf("Applying remove to %d rows", matching.Length())
		matching.Remove()
	}
	if r.definition.Search.Rows.Selector == "" {
		return errors.New("no result item selector is given")
	}
	return nil
}

func (r *Runner) extractSearchTarget(query *search.Query, localCats []string, context RunContext) (*source.SearchTarget, error) {
	// Exposed fields to add:
	templateData := r.getSearchTemplateData(query, localCats, context)
	// ApplyTo our context to the search path
	searchURL, err := templateData.ApplyTo("search_path", r.definition.Search.Path)
	if err != nil {
		return nil, err
	}
	// Resolve the search url
	urlContext, _ := r.GetURLContext()
	searchURL, err = urlContext.GetFullURL(searchURL)
	if err != nil {
		return nil, err
	}
	r.logger.
		WithFields(logrus.Fields{"query": query.Encode()}).
		Debugf("Searching Indexer")
	// Get our Indexer url values
	vals, err := r.extractURLValues(templateData)
	if err != nil {
		return nil, err
	}
	target := &source.SearchTarget{URL: searchURL, Values: vals, Method: r.definition.Search.Method}
	return target, nil
}

func (r *Runner) extractURLValues(templateData *SearchTemplateData) (url.Values, error) {
	// Parse the values that will be used in the url for the search
	urlValues := url.Values{}
	for inputName, inputValueFromScheme := range r.definition.Search.Inputs {
		if templateData.HasQueryField(inputName) {
			inputValueFromScheme = templateData.ApplyField(inputName)
		}
		resolveInputValue, err := templateData.ApplyTo("search_inputs", inputValueFromScheme)
		if err != nil {
			return nil, err
		}
		switch inputName {
		case "$raw":
			err := evalRawSearchInputs(resolveInputValue, r, urlValues)
			if err != nil {
				return urlValues, err
			}
		default:
			urlValues.Add(inputName, resolveInputValue)
		}
	}
	return urlValues, nil
}

func evalRawSearchInputs(resolvedInputValue string, r *Runner, vals url.Values) error {
	parsedVals, err := url.ParseQuery(resolvedInputValue)
	if err != nil {
		r.logger.WithError(err).Warn(err)
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
func (r *Runner) getSearchTemplateData(query *search.Query, localCats []string, context RunContext) *SearchTemplateData {
	startIndex := int(query.Page) * r.definition.Search.PageSize
	context.Search.SetStartIndex(r, startIndex)
	data := newSearchTemplateData(query, localCats, context)
	return data
}

func (r *Runner) hasDateHeader() bool {
	return !r.definition.Search.Rows.DateHeaders.IsEmpty()
}

func (r *Runner) extractDateHeader(selection RawScrapeItem) (time.Time, error) {
	dateHeaders := r.definition.Search.Rows.DateHeaders

	r.logger.
		WithFields(logrus.Fields{"selector": dateHeaders.String()}).
		Debugf("Searching for date header")

	prev := selection.PrevAllFiltered(dateHeaders.Selector).First()
	if prev.Length() == 0 {
		return time.Time{}, fmt.Errorf("no date header row found")
	}

	dv, _ := dateHeaders.Text(prev.First())
	return parseFuzzyTime(dv, time.Now(), true)
}

func (r *Runner) Ratio() (string, error) {
	if r.definition.Ratio.TextVal != "" {
		return r.definition.Ratio.TextVal, nil
	}

	if r.definition.Ratio.Path == "" {
		return "unknown", nil
	}

	r.createBrowser()
	if !r.keepSessions {
		defer r.releaseBrowser()
	}
	urlContext, _ := r.GetURLContext()

	if err := r.session.setup(); err != nil {
		return errorValue, err
	}

	ratioURL, err := urlContext.GetFullURL(r.definition.Ratio.Path)
	if err != nil {
		return errorValue, err
	}

	resultData, err := r.contentFetcher.Fetch(source.NewTarget(ratioURL))
	if err != nil {
		r.logger.WithError(err).Warn("Failed to open page")
		return errorValue, nil
	}

	var ratio interface{}
	switch value := resultData.(type) {
	case *web.HTMLFetchResult:
		ratio, err := r.definition.Ratio.Match(&DomScrapeItem{value.Dom.First()})
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
	errs := make([]string, r.errors.Len())
	for i := 0; i < r.errors.Len(); i++ {
		err, ok := r.errors.Get(i)
		if !ok {
			continue
		}
		errs[i] = fmt.Sprintf("%s", err)
	}
	return errs
}

type StatusReporter struct {
	context         context.Context
	indexDefinition *Definition
	errors          cache.LRUCache
}

func (r *StatusReporter) Error(err error) {
	if err == nil {
		return
	}
	status.PublishSchemeError(r.context, generateSchemeErrorStatus(status.LoginError, err, r.indexDefinition))
	errorID := r.errors.Len()
	r.errors.Add(errorID, err)
}

func (r *Runner) noteError(err error) {
	r.statusReporter.Error(err)
}

// region Status messages

func generateSchemeOkStatus(definition *Definition, runCtx RunContext) *status.ScrapeSchemeMessage {
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
