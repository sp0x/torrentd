package indexer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/sp0x/rutracker-rss/config"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"github.com/sp0x/rutracker-rss/torznab"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"

	"github.com/PuerkitoBio/goquery"
	imdbscraper "github.com/cardigann/go-imdb-scraper"
	"github.com/cardigann/releaseinfo"
	"github.com/f2prateek/train"
	trainlog "github.com/f2prateek/train/log"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf"
	"github.com/sp0x/surf/agent"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"

	"github.com/mrobinsn/go-tvmaze/tvmaze"
)

//"github.com/tehjojo/go-tvmaze/tvmaze"

var (
	_ torznab.Indexer = &Runner{}
)

type RunnerOpts struct {
	Config     config.Config
	CachePages bool
	Transport  http.RoundTripper
}

//Runner works with indexers and their definitions
type Runner struct {
	definition  *IndexerDefinition
	browser     browser.Browsable
	cookies     http.CookieJar
	opts        RunnerOpts
	logger      logrus.FieldLogger
	caps        torznab.Capabilities
	browserLock sync.Mutex
}

type RunContext struct {
	Search search.Search
}

func NewRunner(def *IndexerDefinition, opts RunnerOpts) *Runner {
	logger := logrus.New()
	logger.Level = logrus.InfoLevel
	return &Runner{
		opts:       opts,
		definition: def,
		logger:     logger.WithFields(logrus.Fields{"site": def.Site}),
	}
}

func (r *Runner) createTransport() (http.RoundTripper, error) {
	var t http.Transport
	var custom bool

	if proxyAddr, isset := os.LookupEnv("SOCKS_PROXY"); isset {
		r.logger.
			WithFields(logrus.Fields{"addr": proxyAddr}).
			Debugf("Using SOCKS5 proxy")

		dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("can't connect to the proxy %s: %v", proxyAddr, err)
		}

		t.Dial = dialer.Dial
		custom = true
	}

	if _, isset := os.LookupEnv("TLS_INSECURE"); isset {
		t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		custom = true
	}

	if !custom {
		return &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		}, nil
	}

	return &t, nil
}

func (r *Runner) createBrowser() {
	r.browserLock.Lock()

	if r.cookies == nil {
		r.cookies = jar.NewMemoryCookies()
	}

	bow := surf.NewBrowser()
	bow.SetUserAgent(agent.Chrome())
	bow.SetEncoding(r.definition.Encoding)
	bow.SetAttribute(browser.SendReferer, true)
	bow.SetAttribute(browser.MetaRefreshHandling, true)
	bow.SetCookieJar(r.cookies)
	//bow.SetTimeout(time.Second * 10)

	transport, err := r.createTransport()
	if err != nil {
		panic(err)
	}

	if r.opts.Transport != nil {
		transport = r.opts.Transport
	}

	switch os.Getenv("DEBUG_HTTP") {
	case "1", "true", "basic":
		bow.SetTransport(train.TransportWith(transport, trainlog.New(os.Stderr, trainlog.Basic)))
	case "body":
		bow.SetTransport(train.TransportWith(transport, trainlog.New(os.Stderr, trainlog.Body)))
	case "":
		bow.SetTransport(transport)
	default:
		panic("Unknown value for DEBUG_HTTP")
	}

	r.browser = bow
}

func (r *Runner) releaseBrowser() {
	r.browser = nil
	r.browserLock.Unlock()
}

// checks that the runner has the config values it needs
func (r *Runner) checkHasConfig() error {
	for _, setting := range r.definition.Settings {
		_, ok, err := r.opts.Config.GetSiteOption(r.definition.Site, setting.Name)
		if err != nil {
			return fmt.Errorf("Error reading config for %s: %v", setting.Name, err)
		}
		if !ok {
			return fmt.Errorf("No value for %s.%s in config", r.definition.Site, setting.Name)
		}
	}
	return nil
}

//Get a working url for the indexer
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
	r.logger.WithField("url", u).Debugf("Checking connectivity to url")

	err := r.browser.Open(u)
	if err != nil {
		r.logger.WithError(err).Warn("URL check failed")
		return false
	} else if r.browser.StatusCode() != http.StatusOK {
		r.logger.Warn("URL returned non-ok status")
		return false
	}

	return true
}

//resolvePath resolve a relative url based on the working indexer base url
func (r *Runner) resolvePath(urlPath string) (string, error) {
	if strings.HasPrefix(urlPath, "magnet:") {
		return urlPath, nil
	}
	//Get the base url of the indexer
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
	r.logger.
		WithFields(logrus.Fields{"base": base.String(), "u": resolved.String()}).
		Debugf("Resolving url")

	return resolved.String(), nil
}

// this should eventually upstream into surf browser
func (r *Runner) handleMetaRefreshHeader() error {
	h := r.browser.ResponseHeaders()

	if refresh := h.Get("Refresh"); refresh != "" {
		if s := regexp.MustCompile(`\s*;\s*`).Split(refresh, 2); len(s) == 2 {
			r.logger.
				WithField("fields", s).
				Debug("Found refresh header")

			u, err := r.resolvePath(strings.TrimPrefix(s[1], "url="))
			if err != nil {
				return err
			}

			return r.openPage(u)
		}
	}
	return nil
}

func (r *Runner) openPage(u string) error {
	r.logger.WithField("url", u).Debug("Opening page")

	err := r.browser.Open(u)
	if err != nil {
		return err
	}

	_ = r.cachePage()

	r.logger.
		WithFields(logrus.Fields{"code": r.browser.StatusCode(), "page": r.browser.Url()}).
		Debugf("Finished request")

	if err = r.handleMetaRefreshHeader(); err != nil {
		return err
	}

	return nil
}

func (r *Runner) postToPage(u string, vals url.Values) error {
	r.logger.
		WithFields(logrus.Fields{"url": u, "vals": vals}).
		Debugf("Posting to page")

	if err := r.browser.PostForm(u, vals); err != nil {
		return err
	}

	_ = r.cachePage()

	r.logger.
		WithFields(logrus.Fields{"code": r.browser.StatusCode(), "page": r.browser.Url()}).
		Debugf("Finished request")

	if err := r.handleMetaRefreshHeader(); err != nil {
		return err
	}

	return nil
}

//If caching is enabled, we cache the page's contents in our pagecache
func (r *Runner) cachePage() error {
	if !r.opts.CachePages {
		return nil
	}

	dir := config.GetCachePath(r.definition.Site)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	tmpfile, err := ioutil.TempFile(dir, "pagecache")
	if err != nil {
		r.logger.Warn(err)
		return err
	}

	body := strings.NewReader(r.browser.Body())
	_, _ = io.Copy(tmpfile, body)
	if err = tmpfile.Close(); err != nil {
		return err
	}

	newFile := tmpfile.Name() + ".html"
	_ = os.Rename(tmpfile.Name(), newFile)

	r.logger.
		WithFields(logrus.Fields{"file": "file://" + newFile}).
		Debugf("Wrote page output to cache")

	return nil
}

func (r *Runner) loginViaForm(loginURL, formSelector string, vals map[string]string) error {
	r.logger.
		WithFields(logrus.Fields{"url": loginURL, "form": formSelector, "vals": vals}).
		Debugf("Filling and submitting login form")

	if err := r.openPage(loginURL); err != nil {
		return err
	}

	fm, err := r.browser.Form(formSelector)
	if err != nil {
		return err
	}

	for name, value := range vals {
		r.logger.
			WithFields(logrus.Fields{"key": name, "form": formSelector, "val": value}).
			Debugf("Filling input of form")

		if err = fm.Input(name, value); err != nil {
			r.logger.WithError(err).Error("Filling input failed")
			return err
		}
	}

	r.logger.Debug("Submitting login form")
	defer r.cachePage()

	if err = fm.Submit(); err != nil {
		r.logger.WithError(err).Error("Login failed")
		return err
	}

	r.logger.
		WithFields(logrus.Fields{"code": r.browser.StatusCode(), "page": r.browser.Url()}).
		Debugf("Submitted login form")

	return nil
}

func (r *Runner) loginViaPost(loginURL string, vals map[string]string) error {
	data := url.Values{}
	for key, value := range vals {
		data.Add(key, value)
	}

	return r.postToPage(loginURL, data)
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

//extractInputLogins gets the configured input fields and vals for the login.
func (r *Runner) extractInputLogins() (map[string]string, error) {
	result := map[string]string{}
	//Get configuration for the indexer so we can login
	cfg, err := r.opts.Config.GetSite(r.definition.Site)
	if err != nil {
		return nil, err
	}

	ctx := struct {
		Config map[string]string
	}{
		cfg,
	}

	for name, val := range r.definition.Login.Inputs {
		resolved, err := applyTemplate("login_inputs", val, ctx)
		if err != nil {
			return nil, err
		}

		r.logger.
			WithFields(logrus.Fields{"key": name, "val": resolved}).
			Debugf("Resolved login input template")

		result[name] = resolved
	}

	return result, nil
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

	if p.Path != "" {
		testUrl, err := r.resolvePath(p.Path)
		if err != nil {
			return false, err
		}

		err = r.openPage(testUrl)
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
		r.logger.Debug(r.browser.Body())
		r.logger.
			WithFields(logrus.Fields{"selector": p.Selector}).
			Debug("Selector didn't match page")
		return false, nil
	}

	return true, nil
}

//isLoginRequired Checks if login is required for the given indexer
func (r *Runner) isLoginRequired() (bool, error) {
	if r.definition.Login.IsEmpty() {
		return false, nil
	} else if r.definition.Login.Test.IsEmpty() {
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

func (r *Runner) login() error {
	if r.browser == nil {
		r.createBrowser()
		defer r.releaseBrowser()
	}

	filterLogger = r.logger

	loginUrl, err := r.resolvePath(r.definition.Login.Path)
	if err != nil {
		return err
	}

	vals, err := r.extractInputLogins()
	if err != nil {
		return err
	}

	switch r.definition.Login.Method {
	case "", loginMethodForm:
		if err = r.loginViaForm(loginUrl, r.definition.Login.FormSelector, vals); err != nil {
			return err
		}
	case loginMethodPost:
		if err = r.loginViaPost(loginUrl, vals); err != nil {
			return err
		}
	case loginMethodCookie:
		if err = r.loginViaCookie(loginUrl, vals["cookie"]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown login method %q", r.definition.Login.Method)
	}

	if len(r.definition.Login.Error) > 0 {
		if err = r.definition.Login.hasError(r.browser); err != nil {
			r.logger.WithError(err).Error("Failed to login")
			return err
		}
	}

	match, err := r.matchPageTestBlock(r.definition.Login.Test)
	if err != nil {
		return err
	} else if !match {
		return errors.New("Login check after login failed")
	}

	r.logger.Debug("Successfully logged in")
	return nil
}

func (r *Runner) Info() torznab.Info {
	return torznab.Info{
		ID:       r.definition.Site,
		Title:    r.definition.Name,
		Language: r.definition.Language,
		Link:     r.definition.Links[0],
	}
}

//Capabilities gets the torznab formatted capabilities of this indexer.
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

type extractedItem struct {
	search.ResultItem
	LocalCategoryID   string
	LocalCategoryName string
	LocalId           string
}

// localCategories returns a slice of local categories that should be searched
func (r *Runner) localCategories(query torznab.Query) []string {
	localCats := []string{}
	set := make(map[string]struct{})
	if len(query.Categories) > 0 {
		queryCats := torznab.AllCategories.Subset(query.Categories...)

		// resolve query categories to the exact local, or the local based on parent cat
		for _, id := range r.definition.Capabilities.CategoryMap.ResolveAll(queryCats...) {
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

func (r *Runner) resolveQuery(query torznab.Query) (torznab.Query, error) {
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
		movie, err = imdbscraper.FindByID(query.IMDBID)
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

func (r *Runner) GetEncoding() string {
	return r.definition.Encoding
}

//Search for a given torrent
func (r *Runner) Search(query torznab.Query) (*search.Search, error) {
	r.createBrowser()
	defer r.releaseBrowser()
	//[]torznab.ResultItem
	//Search

	var err error
	query, err = r.resolveQuery(query)
	if err != nil {
		return nil, err
	}

	// TODO: make this concurrency safe
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
	//Get the categories for this query based on the indexer
	localCats := r.localCategories(query)

	r.logger.Debugf("Query is %v\n", query)
	r.logger.Debugf("Keywords are %q\n", query.Keywords())
	//Context about the search
	context := RunContext{}

	//Exposed fields to add:
	templateCtx := r.getRunnerContext(query, localCats, context)
	//Apply our context to the search path

	searchURL, err := applyTemplate("search_path", r.definition.Search.Path, templateCtx)
	if err != nil {
		return nil, err
	}
	//Resolve the search url
	searchURL, err = r.resolvePath(searchURL)
	if err != nil {
		return nil, err
	}

	r.logger.
		WithFields(logrus.Fields{"query": query.Encode()}).
		Infof("Searching indexer")

	vals := url.Values{}
	//Parse the values that will be used in the url for the search
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

	timer := time.Now()
	//Get the content
	err = r.requireContent(vals, searchURL)
	if err != nil {
		return nil, err
	}
	dom := r.browser.Dom()
	setupContext(r, &context, dom)
	// merge following rows for After selector
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
			WithFields(logrus.Fields{"selector": remove}).
			Debugf("Applying remove to %d rows", matching.Length())
		matching.Remove()
	}
	if r.definition.Search.Rows.Selector == "" {
		return nil, errors.New("no result item selector is given")
	}
	rows := dom.Find(r.definition.Search.Rows.Selector)

	r.logger.
		WithFields(logrus.Fields{
			"rows":     rows.Length(),
			"selector": r.definition.Search.Rows.Selector,
			"limit":    query.Limit,
			"offset":   query.Offset,
		}).Debugf("Found %d rows", rows.Length())

	var extracted []extractedItem

	for i := 0; i < rows.Length(); i++ {
		if query.Limit > 0 && len(extracted) >= query.Limit {
			break
		}

		item, err := r.extractItem(i+1, rows.Eq(i))
		if err != nil {
			return nil, err
		}

		var matchCat bool
		if len(localCats) > 0 {
			for _, catId := range localCats {
				if catId == item.LocalCategoryID {
					matchCat = true
				}
			}

			if !matchCat {
				r.logger.
					WithFields(logrus.Fields{"category": item.LocalCategoryName, "categoryId": item.LocalCategoryID}).
					Debugf("Skipping result because it's not contained in our needed categories.")
				continue
			}
		}
		//Try to map the category from the indexer to the global categories
		r.resolveCategory(&item)
		if query.Series != "" {
			info, err := releaseinfo.Parse(item.Title)
			if err != nil {
				r.logger.
					WithFields(logrus.Fields{"title": item.Title}).
					WithError(err).
					Warn("Failed to parse show title, skipping")
				continue
			}

			if info != nil && !info.SeriesTitleInfo.Equal(query.Series) {
				r.logger.
					WithFields(logrus.Fields{"got": info.SeriesTitleInfo.TitleWithoutYear, "expected": query.Series}).
					Debugf("Series search skipping non-matching series")
				continue
			}
		}
		extracted = append(extracted, item)
	}

	r.logger.
		WithFields(logrus.Fields{"time": time.Now().Sub(timer)}).
		Infof("Query returned %d results", len(extracted))

	var items []search.ResultItem
	for _, item := range extracted {
		items = append(items, item.ResultItem)
	}
	srchResult := &context.Search
	srchResult.Results = items
	return srchResult, nil
}

type RunnerPatternData struct {
	Query      torznab.Query
	Keywords   string
	Categories []string
	Context    RunContext
}

func (r *Runner) getRunnerContext(query torznab.Query, localCats []string, context RunContext) RunnerPatternData {
	templateCtx := RunnerPatternData{
		query,
		query.Keywords(),
		localCats,
		context,
	}
	return templateCtx
}

//Gets the content from which we'll extract the search results
func (r *Runner) requireContent(urlVals url.Values, searchURL string) error {
	var err error
	switch r.definition.Search.Method {
	case "", searchMethodGet:
		if len(urlVals) > 0 {
			searchURL = fmt.Sprintf("%s?%s", searchURL, urlVals.Encode())
		}
		if err = r.openPage(searchURL); err != nil {
			return err
		}
	case searchMethodPost:
		if err = r.postToPage(searchURL, urlVals); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown search method %q", r.definition.Search.Method)
	}
	return nil
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
	return parseFuzzyTime(dv, time.Now())
}

func (r *Runner) Download(u string) (io.ReadCloser, http.Header, error) {
	r.createBrowser()

	if required, err := r.isLoginRequired(); required {
		if err := r.login(); err != nil {
			r.logger.WithError(err).Error("Login failed")
			return nil, http.Header{}, err
		}
	} else if err != nil {
		return nil, http.Header{}, err
	}

	fullUrl, err := r.resolvePath(u)
	if err != nil {
		return nil, http.Header{}, err
	}

	if err := r.browser.Open(fullUrl); err != nil {
		return nil, http.Header{}, err
	}

	pipeR, pipeW := io.Pipe()
	go func() {
		defer pipeW.Close()
		defer r.releaseBrowser()
		n, err := r.browser.Download(pipeW)
		if err != nil {
			r.logger.Error(err)
		}
		r.logger.WithFields(logrus.Fields{"url": fullUrl}).Debugf("Downloaded %d bytes", n)
	}()

	return pipeR, r.browser.ResponseHeaders(), nil
}

func (r *Runner) Ratio() (string, error) {
	if r.definition.Ratio.TextVal != "" {
		return r.definition.Ratio.TextVal, nil
	}

	if r.definition.Ratio.Path == "" {
		return "unknown", nil
	}

	r.createBrowser()
	defer r.releaseBrowser()

	if required, err := r.isLoginRequired(); required {
		if err := r.login(); err != nil {
			r.logger.WithError(err).Error("Login failed")
			return "error", err
		}
	} else if err != nil {
		return "error", err
	}

	ratioUrl, err := r.resolvePath(r.definition.Ratio.Path)
	if err != nil {
		return "error", err
	}

	err = r.openPage(ratioUrl)
	if err != nil {
		r.logger.WithError(err).Warn("Failed to open page")
		return "error", nil
	}

	ratio, err := r.definition.Ratio.MatchText(r.browser.Dom())
	if err != nil {
		return ratio, err
	}

	return strings.Trim(ratio, "- "), nil
}

func (r *Runner) getIndexer() *search.ResultIndexer {
	return &search.ResultIndexer{
		Id:   "",
		Name: r.definition.Site,
	}
}
