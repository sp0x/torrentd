package indexer

import (
	"errors"
	"fmt"
	"github.com/sp0x/torrentd/indexer/source/series"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/categories"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/source"
	"github.com/sp0x/torrentd/storage"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/sp0x/torrentd/torznab"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	imdbscraper "github.com/cardigann/go-imdb-scraper"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"

	"github.com/mrobinsn/go-tvmaze/tvmaze"
)

//"github.com/tehjojo/go-tvmaze/tvmaze"

var (
	_ Indexer = &Runner{}
)

type RunnerOpts struct {
	Config     config.Config
	CachePages bool
	Transport  http.RoundTripper
}

//Runner works index definitions in order to extract data.
type Runner struct {
	definition *IndexerDefinition
	browser    browser.Browsable
	cookies    http.CookieJar
	opts       RunnerOpts
	logger     logrus.FieldLogger
	//caps                torznab.Capabilities
	browserLock         sync.Mutex
	connectivityTester  cache.ConnectivityTester
	state               *IndexerState
	keepSessions        bool
	failingSearchFields map[string]fieldBlock
	lastVerified        time.Time
	Storage             storage.ItemStorage
	contentFetcher      source.ContentFetcher
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
	//SearchKeywords *search.Instance
}

//NewRunner Start a runner for a given indexer.
func NewRunner(def *IndexerDefinition, opts RunnerOpts) *Runner {
	logger := logrus.New()
	logger.Level = logrus.GetLevel()
	connCache, _ := cache.NewConnectivityCache()
	runnerConfig := opts.Config
	runner := &Runner{
		opts:                opts,
		definition:          def,
		logger:              logger.WithFields(logrus.Fields{"site": def.Site}),
		connectivityTester:  connCache,
		state:               defaultIndexerState(),
		keepSessions:        true,
		failingSearchFields: make(map[string]fieldBlock),
	}
	entityType := runner.definition.getSearchEntity()
	storageType := opts.Config.GetString("storage")
	if storageType == "" {
		panic("no storage type configured")
	}
	if entityType != nil {
		//All the results will be stored in a collection with the same name as the index.
		//runner.Storage = storage.NewKeyedStorageWithBackingType(def.Name, runnerConfig, entityType.GetKey(), storageType)
		runner.Storage = storage.NewBuilder().
			WithNamespace(def.Name).
			WithEndpoint(runnerConfig.GetString("db")).
			WithPK(entityType.GetKey()).
			WithBacking(storageType).
			WithRecord(&search.ExternalResultItem{}).
			Build()
	} else {
		//All the results will be stored in a collection with the same name as the index.
		//runner.Storage = storage.NewKeyedStorageWithBackingType(def.Name, runnerConfig, indexing.NewKey(), storageType)
		runner.Storage = storage.NewBuilder().
			WithNamespace(def.Name).
			WithEndpoint(runnerConfig.GetString("db")).
			WithBacking(storageType).
			WithRecord(&search.ExternalResultItem{}).
			Build()
	}

	return runner
}

// checks that the runner has the config values it needs
//func (r *Runner) checkHasConfig() error {
//	for _, setting := range r.definition.Settings {
//		_, ok, err := r.opts.Config.GetSiteOption(r.definition.Site, setting.Name)
//		if err != nil {
//			return fmt.Errorf("Error reading config for %s: %v", setting.Name, err)
//		}
//		if !ok {
//			return fmt.Errorf("No value for %s.%s in config", r.definition.Site, setting.Name)
//		}
//	}
//	return nil
//}

//Get a working url for the Indexer
func (r *Runner) currentURL() (*url.URL, error) {
	if u := r.browser.Url(); u != nil {
		return u, nil
	}
	configURL, ok, _ := r.opts.Config.GetSiteOption(r.definition.Site, "url")
	if ok && r.testURLWorks(configURL) {
		return url.Parse(configURL)
	}
	for _, u := range r.definition.Links {
		if u != configURL && r.testURLWorks(u) {
			return url.Parse(u)
		}
	}
	return nil, errors.New("No working urls found")
}

func (r *Runner) testURLWorks(u string) bool {
	var ok bool
	//Do this like that so it's locked.
	if ok = r.connectivityTester.IsOkAndSet(u, func() bool {
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

//resolveIndexerPath resolve a relative url based on the working Indexer base url
func (r *Runner) resolveIndexerPath(urlPath string) (string, error) {
	if strings.HasPrefix(urlPath, "magnet:") {
		return urlPath, nil
	}
	//Get the base url of the Indexer
	base, err := r.currentURL()
	if err != nil {
		return "", err
	}

	u, err := url.Parse(urlPath)
	if err != nil {
		return "", err
	}
	//Resolve the url
	resolved := base.ResolveReference(u)
	return resolved.String(), nil
}

func parseCookieString(cookie string) []*http.Cookie {
	h := http.Header{"Cookie": []string{cookie}}
	r := http.Request{Header: h}
	return r.Cookies()
}

func (r *Runner) loginViaCookie(loginURL string, cookie string) error {
	u, err := url.Parse(loginURL)
	if err != nil {
		return err
	}

	cookies := parseCookieString(cookie)

	r.logger.
		WithFields(logrus.Fields{"url": loginURL, "cookies": cookies}).
		Debugf("Setting cookies for login")

	cj := jar.NewMemoryCookies()
	cj.SetCookies(u, cookies)

	r.browser.SetCookieJar(cj)
	return nil
}

func (r *Runner) matchPageTestBlock(p pageTestBlock) (bool, error) {
	if p.IsEmpty() {
		return true, nil
	}

	r.logger.
		WithFields(logrus.Fields{"path": p.Path, "selector": p.Selector}).
		Debug("Checking page test block")

	if r.browser.Url() == nil && p.Path == "" {
		return false, errors.New("No url loaded and pageTestBlock has no path")
	}
	//Go to a path to verify
	if p.Path != "" {
		testUrl, err := r.resolveIndexerPath(p.Path)
		if err != nil {
			return false, err
		}

		err = r.contentFetcher.Fetch(source.NewTarget(testUrl))
		if err != nil {
			r.logger.WithError(err).Warn("Failed to open page")
			return false, nil
		}

		if testUrl != r.browser.Url().String() {
			r.logger.
				WithFields(logrus.Fields{"wanted": testUrl, "got": r.browser.Url().String()}).
				Debug("Test failed, got a redirect")
			return false, nil
		}
	}

	if p.Selector != "" && r.browser.Find(p.Selector).Length() == 0 {
		body := r.browser.Body()
		r.logger.Debug(body)
		r.logger.
			WithFields(logrus.Fields{"selector": p.Selector}).
			Debug("Selector didn't match page")
		return false, nil
	}

	return true, nil
}

//isLoginRequired Checks if login is required for the given Indexer
func (r *Runner) isLoginRequired() (bool, error) {
	if r.definition.Login.IsEmpty() {
		return false, nil
	} else if r.definition.Login.Test.IsEmpty() {
		return true, nil
	}
	isLoggedIn := r.state.GetBool("loggedIn")
	if !isLoggedIn {
		return true, nil
	}
	r.logger.Debug("Testing if login is needed")
	//Check if the login page is valid
	match, err := r.matchPageTestBlock(r.definition.Login.Test)
	if err != nil {
		return true, err
	}

	if match {
		r.logger.Debug("No login needed, already logged in")
		return false, nil
	}

	r.logger.Debug("Login is required")
	return true, nil
}

//Capabilities gets the torznab formatted capabilities of this Indexer.
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

// getLocalCategoriesMatchingQuery returns a slice of local categories that should be searched
func (r *Runner) getLocalCategoriesMatchingQuery(query *torznab.Query) []string {
	var localCats []string
	set := make(map[string]struct{})
	if len(query.Categories) > 0 {
		queryCats := categories.AllCategories.Subset(query.Categories...)
		// resolve query categories to the exact local, or the local based on parent cat
		for _, id := range r.definition.Capabilities.CategoryMap.ResolveAll(queryCats.Items()...) {
			//Add only if it doesn't exist
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

func (r *Runner) fillInAdditionalQueryParameters(query *torznab.Query) (*torznab.Query, error) {
	var show *tvmaze.Show
	var movie *imdbscraper.Movie
	var err error

	// convert show identifiers to season parameter
	switch {
	case query.TVDBID != "" && query.TVDBID != "0":
		show, err = tvmaze.DefaultClient.GetShowWithTVDBID(query.TVDBID)
		query.TVDBID = "0"
	case query.TVMazeID != "":
		show, err = tvmaze.DefaultClient.GetShowWithID(query.TVMazeID)
		query.TVMazeID = "0"
	case query.TVRageID != "":
		show, err = tvmaze.DefaultClient.GetShowWithTVRageID(query.TVRageID)
		query.TVRageID = ""
	case query.IMDBID != "":
		imdbid := query.IMDBID
		if !strings.HasPrefix(imdbid, "tt") {
			imdbid = "tt" + imdbid
		}
		movie, err = imdbscraper.FindByID(imdbid)
		if err != nil {
			err = fmt.Errorf("imdb error. %s", err)
		}
		query.IMDBID = ""
	}

	if err != nil {
		return query, err
	}

	if show != nil {
		query.Series = show.Name
		r.logger.
			WithFields(logrus.Fields{"name": show.Name, "year": show.GetFirstAired().Year()}).
			Debugf("Found show via tvmaze lookup")
	}

	if movie != nil {
		if movie.Title == "" {
			return query, fmt.Errorf("Movie title was blank")
		}
		query.Movie = movie.Title
		query.Year = movie.Year
		r.logger.
			WithFields(logrus.Fields{"title": movie.Title, "year": movie.Year, "movie": movie}).
			Debugf("Found movie via imdb lookup")

	}

	return query, nil
}

//GetEncoding returns the encoding that's set to be used in this index.
//This can be changed in the index's definition.
func (r *Runner) GetEncoding() string {
	return r.definition.Encoding
}

//Check sees if the index can be searched
func (r *Runner) Check() error {
	verifiedSpan := time.Since(r.lastVerified)
	if verifiedSpan < time.Minute*60*24 {
		return nil
	}
	_, err := r.Search(&torznab.Query{}, nil)
	if err != nil {
		r.lastVerified = time.Now()
	}
	return err
}

func (r *Runner) getUniqueIndex(item *search.ExternalResultItem) *indexing.Key {
	if item == nil {
		return nil
	}
	key := indexing.NewKey()
	//Local id would be a good bet.
	if len(item.LocalId) > 0 {
		key.Add("LocalId")
	}
	return key
}

//SearchKeywords for a given torrent
func (r *Runner) Search(query *torznab.Query, srch search.Instance) (search.Instance, error) {
	r.createBrowser()
	if !r.keepSessions {
		defer r.releaseBrowser()
	}
	var err error
	query, err = r.fillInAdditionalQueryParameters(query)
	if err != nil {
		return nil, err
	}
	filterLogger = r.logger
	//Login if it's required
	if required, err := r.isLoginRequired(); err != nil {
		return nil, err
	} else if required {
		if err := r.login(); err != nil {
			r.logger.WithError(err).Error("Login failed")
			return nil, err
		}
	}
	//Get the categories for this query based on the Indexer
	localCats := r.getLocalCategoriesMatchingQuery(query)

	//Context about the search
	if srch == nil {
		srch = &search.Search{}
	}
	context := RunContext{
		Search: srch.(*search.Search),
	}
	target, err := r.extractSearchTarget(query, localCats, context)
	if err != nil {
		return nil, err
	}
	timer := time.Now()
	//Get the content
	err = r.contentFetcher.Fetch(target)
	if err != nil {
		return nil, err
	}
	r.logger.
		WithFields(logrus.Fields{}).
		Debugf("Fetched Indexer page.\n")
	dom := r.browser.Dom()
	setupContext(r, &context, dom)
	// merge following rows for After selector
	err = r.clearDom(dom)
	if err != nil {
		return nil, err
	}
	rows := dom.Find(r.definition.Search.Rows.Selector)
	r.logger.
		WithFields(logrus.Fields{
			"rows":     rows.Length(),
			"selector": r.definition.Search.Rows.Selector,
			"limit":    query.Limit,
			"offset":   query.Offset,
		}).Debugf("Found %d rows", rows.Length())

	var results []search.ExternalResultItem
	for i := 0; i < rows.Length(); i++ {
		if query.Limit > 0 && len(results) >= query.Limit {
			break
		}
		//Get the result from the row
		item, err := r.extractItem(i+1, rows.Eq(i))
		if err != nil {
			continue
		}
		//Maybe don't do that always?
		item.Fingerprint = search.GetResultFingerprint(&item)
		if !r.validateAndStoreItem(query, localCats, &item) {
			_ = r.Storage.SetKey(r.getUniqueIndex(&item))
			err = r.Storage.Add(&item)
			if err != nil {
				r.logger.Errorf("Found an item that doesn't match our search categories: %s\n", err)
			}
			continue
		}
		_ = r.Storage.SetKey(r.getUniqueIndex(&item))
		err = r.Storage.Add(&item)
		if err != nil {
			r.logger.Errorf("Couldn't add item: %s\n", err)
		}
		results = append(results, item)
	}
	r.logger.
		WithFields(logrus.Fields{"Indexer": r.definition.Site, "q": query.Keywords(), "time": time.Since(timer)}).
		Infof("Query returned %d results", len(results))
	context.Search.SetResults(results)
	return context.Search, nil
}

func (r *Runner) validateAndStoreItem(query *torznab.Query, localCats []string, item *search.ExternalResultItem) bool {
	if len(localCats) > 0 {
		//The category doesn't match even 1 of the categories in the query.
		if !r.itemMatchesLocalCategories(localCats, item) {
			r.logger.
				WithFields(logrus.Fields{"category": item.LocalCategoryName, "categoryId": item.LocalCategoryID}).
				Debugf("Skipping result because it's not contained in our needed categories.")
			return false
		}
	}
	//Try to map the category from the Indexer to the global categories
	r.resolveCategory(item)
	return !series.IsSeriesAndNotMatching(query, item)
}

func (r *Runner) itemMatchesLocalCategories(localCats []string, item *search.ExternalResultItem) bool {
	for _, catId := range localCats {
		if catId == item.LocalCategoryID {
			return true
		}
	}
	return false
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

func (r *Runner) extractSearchTarget(query *torznab.Query, localCats []string, context RunContext) (*source.SearchTarget, error) {
	//Exposed fields to add:
	templateCtx := r.getRunnerContext(query, localCats, context)
	//Apply our context to the search path
	searchURL, err := applyTemplate("search_path", r.definition.Search.Path, templateCtx)
	if err != nil {
		return nil, err
	}
	//Resolve the search url
	searchURL, err = r.resolveIndexerPath(searchURL)
	if err != nil {
		return nil, err
	}
	r.logger.
		WithFields(logrus.Fields{"query": query.Encode()}).
		Debugf("Searching Indexer")
	//Get our Indexer url values
	vals, err := r.extractUrlValues(templateCtx)
	if err != nil {
		return nil, err
	}
	target := &source.SearchTarget{Url: searchURL, Values: vals}
	return target, nil
}

func (r *Runner) extractUrlValues(templateCtx RunnerPatternData) (url.Values, error) {
	//Parse the values that will be used in the url for the search
	vals := url.Values{}
	for name, val := range r.definition.Search.Inputs {
		resolved, err := applyTemplate("search_inputs", val, templateCtx)
		if err != nil {
			return nil, err
		}
		switch name {
		case "$raw":
			parsedVals, err := url.ParseQuery(resolved)
			if err != nil {
				r.logger.WithError(err).Warn(err)
				return nil, fmt.Errorf("Error parsing $raw input: %s", err.Error())
			}

			r.logger.
				WithFields(logrus.Fields{"source": val, "parsed": parsedVals}).
				Debugf("Processed $raw input")

			for k, values := range parsedVals {
				for _, val := range values {
					vals.Add(k, val)
				}
			}
		default:
			vals.Add(name, resolved)
		}
	}
	return vals, nil
}

type RunnerPatternData struct {
	Query      *torznab.Query
	Keywords   string
	Categories []string
	Context    RunContext
}

//Get the default run context
func (r *Runner) getRunnerContext(query *torznab.Query, localCats []string, context RunContext) RunnerPatternData {
	startIndex := int(query.Page) * r.definition.Search.PageSize
	context.Search.SetStartIndex(r, startIndex)
	templateCtx := RunnerPatternData{
		query,
		query.Keywords(),
		localCats,
		context,
	}
	return templateCtx
}

func (r *Runner) hasDateHeader() bool {
	return !r.definition.Search.Rows.DateHeaders.IsEmpty()
}

func (r *Runner) extractDateHeader(selection *goquery.Selection) (time.Time, error) {
	dateHeaders := r.definition.Search.Rows.DateHeaders

	r.logger.
		WithFields(logrus.Fields{"selector": dateHeaders.String()}).
		Debugf("Searching for date header")

	prev := selection.PrevAllFiltered(dateHeaders.Selector).First()
	if prev.Length() == 0 {
		return time.Time{}, fmt.Errorf("No date header row found")
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

	if required, err := r.isLoginRequired(); required {
		if err := r.login(); err != nil {
			r.logger.WithError(err).Error("Login failed")
			return "error", err
		}
	} else if err != nil {
		return "error", err
	}

	ratioUrl, err := r.resolveIndexerPath(r.definition.Ratio.Path)
	if err != nil {
		return "error", err
	}

	err = r.contentFetcher.Fetch(source.NewTarget(ratioUrl))
	if err != nil {
		r.logger.WithError(err).Warn("Failed to open page")
		return "error", nil
	}

	ratio, err := r.definition.Ratio.Match(r.browser.Dom())
	if err != nil {
		return ratio.(string), err
	}

	return strings.Trim(ratio.(string), "- "), nil
}

func (r *Runner) getIndexer() *search.ResultIndexer {
	return &search.ResultIndexer{
		Id:   "",
		Name: r.definition.Site,
	}
}

//func (r *Runner) getField(s string) *fieldBlock {
//	for _, fld := range r.definition.Search.Fields {
//		if fld.Field == s {
//			return &fld
//		}
//	}
//	return nil
//}
