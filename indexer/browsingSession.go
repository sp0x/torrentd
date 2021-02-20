package indexer

import (
	"errors"
	"fmt"
	"github.com/sp0x/surf/jar"
	"net/url"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer/source"
)

const emptyValue = "<no value>"

type LoginState int

const (
	NoLoginRequired LoginState = iota + 1
	LoginExpired
	LoginRequired
	LoginFailed
	LoggedIn
)

type BrowsingSessionMultiplexer struct {
	sessions []*BrowsingSession
	index    int
}

type BrowsingSession struct {
	loginBlock     loginBlock
	state          LoginState
	urlContext     *URLContext
	contentFetcher *source.Fetcher
	config         map[string]string
	logger         *logrus.Logger
	statusReporter *StatusReporter
}

func NewSessionMultiplexer(runner *Runner, sessionCount int) (*BrowsingSessionMultiplexer, error) {
	mux := &BrowsingSessionMultiplexer{}
	mux.sessions = make([]*BrowsingSession, sessionCount)
	var wg sync.WaitGroup
	wg.Add(sessionCount)
	var err error

	for i := 0; i < sessionCount; i++ {
		go func(index int) {
			defer wg.Done()
			if err != nil {
				return
			}
			session, tmpErr := newIndexSessionFromRunner(runner)
			if tmpErr != nil {
				err = tmpErr
				return
			}
			mux.sessions[index] = session
		}(i)
	}
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return mux, nil
}

func (b *BrowsingSessionMultiplexer) acquire() (*BrowsingSession, error) {
	session := b.sessions[b.index%len(b.sessions)]
	b.index++
	if err := session.setup(); err != nil {
		return nil, err
	} else {
		return session, nil
	}
}

func newIndexSessionFromRunner(runner *Runner) (*BrowsingSession, error) {
	urlContext, err := runner.GetURLContext()
	if err != nil {
		return nil, err
	}
	definition := runner.definition
	webFetcher := createContentFetcher(runner)
	siteConfig, err := runner.options.Config.GetSite(definition.Name)
	if err != nil {
		return nil, err
	}
	browsingSession := newIndexSessionWithLogin(
		siteConfig,
		runner.statusReporter,
		webFetcher,
		urlContext,
		definition.Login)
	return browsingSession, nil
}

func newIndexSessionWithLogin(siteConfig map[string]string,
	statusReporter *StatusReporter,
	contentFetcher *source.Fetcher,
	urlContext *URLContext,
	loginBlock loginBlock) *BrowsingSession {

	lc := &BrowsingSession{}
	lc.loginBlock = loginBlock
	lc.urlContext = urlContext
	lc.contentFetcher = contentFetcher
	lc.config = siteConfig
	lc.logger = logrus.New()
	lc.statusReporter = statusReporter
	if loginBlock.IsEmpty() {
		lc.state = NoLoginRequired
	} else {
		lc.state = LoginRequired
	}
	return lc
}

func (l *BrowsingSession) isRequired() bool {
	if l.state == LoginRequired ||
		l.state == LoginFailed ||
		l.state == LoginExpired {
		return true
	}
	return false
}

func (l *BrowsingSession) verifyLogin() (bool, error) {
	testBlock := l.loginBlock.Test
	if testBlock.IsEmpty() {
		return true, nil
	}
	var loginResult *source.HTMLFetchResult

	// Go to another url if needed
	if testBlock.Path != "" {
		testURL, err := l.urlContext.GetFullURL(testBlock.Path)
		if err != nil {
			return false, err
		}

		r, err := l.contentFetcher.Fetch(source.NewRequestOptions(testURL))
		if htmlRes, ok := r.(*source.HTMLFetchResult); !ok {
			return false, errors.New("expected html from login")
		} else {
			loginResult = htmlRes
		}
		if err != nil {
			// r.logger.WithError(err).Warn("Failed to open page")
			return false, nil
		}
		fetchedAddress := l.contentFetcher.URL().String()

		if testURL != fetchedAddress {
			return false, nil
		}
		if loginResult == nil || loginResult.DOM == nil {
			return false, errors.New("could not get DOM for login")
		}
	}

	if l.contentFetcher.URL() == nil {
		return false, errors.New("no URL loaded and pageTestBlock has no path")
	}

	if testBlock.Selector != "" && l.contentFetcher.Browser.Find(testBlock.Selector).Length() == 0 {
		return false, nil
	}

	return true, nil
}

// extractLoginInput gets the configured input fields and values for the login.
func (l *BrowsingSession) extractLoginInput() (map[string]string, error) {
	result := map[string]string{}
	loginUrl, _ := l.urlContext.GetFullURL(l.loginBlock.Path)

	// Get configuration for the Index so we can login
	ctx := struct {
		Config map[string]string
	}{
		l.config,
	}
	for name, val := range l.loginBlock.Inputs {
		resolved, err := applyTemplate("login_inputs", val, ctx)
		if err != nil {
			return nil, err
		}

		if val == "{{ .Config.password }}" && resolved == emptyValue {
			return nil, fmt.Errorf("no password was configured for input `%s` @%s", name, loginUrl)
		}
		result[name] = resolved
	}

	return result, nil
}

func (l *BrowsingSession) initLogin() error {
	if l.loginBlock.Init.IsEmpty() {
		return nil
	}

	initURL, err := l.urlContext.GetFullURL(l.loginBlock.Init.Path)
	if err != nil {
		return err
	}
	_, err = l.contentFetcher.Fetch(&source.RequestOptions{
		Method: "get",
		URL:    initURL,
	})
	return err
}

func (l *BrowsingSession) login() error {
	loginURL, err := l.urlContext.GetFullURL(l.loginBlock.Path)
	if err != nil {
		return err
	}

	loginValues, err := l.extractLoginInput()
	if err != nil {
		return err
	}

	err = l.initLogin()
	if err != nil {
		return err
	}

	method := l.loginBlock.Method
	switch method {
	case "", loginMethodForm:
		if err = l.loginViaForm(loginURL, l.loginBlock.FormSelector, loginValues); err != nil {
			return err
		}
	case loginMethodPost:
		if err = l.loginViaPost(loginURL, loginValues); err != nil {
			return err
		}
	case loginMethodCookie:
		cookieVal := loginValues["cookie"]
		if cookieVal == emptyValue {
			return &LoginError{errors.New("no login cookie configured")}
		}
		if err = l.loginViaCookie(loginURL, loginValues["cookie"]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown login method %q for site %s", method, loginURL)
	}
	// Get the error
	if len(l.loginBlock.Error) > 0 {
		if err = l.loginBlock.hasError(l.contentFetcher.Browser); err != nil {
			l.logger.WithError(err).Error("Failed to login")
			return &LoginError{err}
		}
	}
	// Check if the login was successful
	loggedIn, err := l.verifyLogin()
	if err != nil {
		return err
	} else if !loggedIn {
		hasPass := loginValues["login_password"] != emptyValue
		if _, ok := loginValues["login_password"]; !ok {
			hasPass = false
		}
		return fmt.Errorf("login check after login failed. no matches found. user: %s, using pass: %v", loginValues["login_username"], hasPass)
	}

	l.state = LoggedIn
	return nil
}

func (l *BrowsingSession) loginViaCookie(loginURL string, cookie string) error {
	u, err := url.Parse(loginURL)
	if err != nil {
		return err
	}

	cookies := parseCookieString(cookie)
	cj := jar.NewMemoryCookies()
	cj.SetCookies(u, cookies)

	l.contentFetcher.Browser.SetCookieJar(cj)
	return nil
}

func (l *BrowsingSession) loginViaForm(loginURL, formSelector string, vals map[string]string) error {
	_, err := l.contentFetcher.Fetch(&source.RequestOptions{Method: "get", URL: loginURL})
	if err != nil {
		return err
	}

	webForm, err := l.contentFetcher.Browser.Form(formSelector)
	if err != nil {
		return err
	}

	for name, value := range vals {
		if err = webForm.Input(name, value); err != nil {
			//r.logger.WithError(err).Error("Filling input failed")
			return err
		}
	}

	// Maybe we don't need to cache the current browser page
	// defer r.cachePage()
	if err = webForm.Submit(); err != nil {
		l.logger.WithError(err).Error("Login failed")
		return err
	}
	//r.logger.
	//	WithFields(logrus.Fields{"code": r.browser.StatusCode(), "page": r.browser.Url()}).
	//	Debugf("Submitted login form")

	return nil
}

func (l *BrowsingSession) loginViaPost(loginURL string, vals map[string]string) error {
	data := url.Values{}
	for key, value := range vals {
		data.Add(key, value)
	}
	options := &source.RequestOptions{
		Method: "post",
		Values: data,
		URL:    loginURL,
	}
	return l.contentFetcher.Post(options)
}

func (l *BrowsingSession) setup() error {
	if !l.isRequired() {
		return nil
	}
	if err := l.login(); err != nil {
		l.logger.WithError(err).Error("Login failed")
		l.statusReporter.Error(err)
		return err
	}
	return nil
}

func (l *BrowsingSession) ApplyToRequest(target *source.RequestOptions) {
	brw := l.contentFetcher.Browser
	cookies := brw.CookieJar()
	target.CookieJar = cookies
	target.Referer = brw.Url()
}
