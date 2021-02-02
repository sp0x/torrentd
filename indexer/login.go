package indexer

import (
	"errors"
	"fmt"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"
	"github.com/sp0x/torrentd/config"
	"net/url"

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

type BrowsingSession struct {
	loginBlock     loginBlock
	testBlock      pageTestBlock
	state          LoginState
	browser        browser.Browsable
	urlContext     *URLContext
	contentFetcher source.ContentFetcher
	config         config.Config
	site           string
	logger         *logrus.Logger
}

func newIndexSessionFromRunner(runner *Runner) (*BrowsingSession, error) {
	urlContext, err := runner.GetURLContext()
	if err != nil {
		return nil, err
	}
	definition := runner.definition
	browsingSession := newIndexSessionWithLogin(definition.Name, runner.options.Config, runner.browser,
		runner.contentFetcher, urlContext, definition.Login)
	return browsingSession, nil
}

func newIndexSessionWithLogin(site string, cfg config.Config, browser browser.Browsable, contentFetcher source.ContentFetcher, urlContext *URLContext, loginBlock loginBlock) *BrowsingSession {
	lc := &BrowsingSession{}
	lc.loginBlock = loginBlock
	lc.testBlock = loginBlock.Test
	lc.urlContext = urlContext
	lc.contentFetcher = contentFetcher
	lc.config = cfg
	lc.site = site
	lc.logger = logrus.New()
	lc.browser = browser
	if loginBlock.IsEmpty() {
		lc.state = NoLoginRequired
	} else {
		lc.state = LoginRequired
	}
	return lc
}

func (l *BrowsingSession) isRequired() bool {
	// r.logger.Debug("Testing if login is needed")
	// HealthCheck if the login page is valid
	if l.state == LoginRequired ||
		l.state == LoginFailed ||
		l.state == LoginExpired {
		return true
	}
	return false
}

func (l *BrowsingSession) verifyLogin() (bool, error) {
	testBlock := l.testBlock
	browser := l.browser
	if testBlock.IsEmpty() {
		return true, nil
	}

	// Go to another url if needed
	if testBlock.Path != "" {
		testURL, err := l.urlContext.GetFullURL(testBlock.Path)
		if err != nil {
			return false, err
		}

		_, err = l.contentFetcher.Fetch(source.NewTarget(testURL))
		if err != nil {
			// r.logger.WithError(err).Warn("Failed to open page")
			return false, nil
		}
		fetchedAddress := l.contentFetcher.URL().String()

		if testURL != fetchedAddress {
			// r.logger.
			//	WithFields(logrus.Fields{"wanted": testURL, "got": r.browser.Url().String()}).
			//	Debug("Test failed, got a redirect")
			return false, nil
		}
	}

	if l.contentFetcher.URL() == nil {
		return false, errors.New("no URL loaded and pageTestBlock has no path")
	}

	if testBlock.Selector != "" && browser.Find(testBlock.Selector).Length() == 0 {
		// body := r.browser.Body()
		// r.logger.Debug(body)
		// r.logger.
		//	WithFields(logrus.Fields{"selector": p.Selector}).
		//	Debug("Selector didn't match page")
		return false, nil
	}

	return true, nil
}

// isLoginRequired Checks if login is required for the given Indexer
//func (r *Runner) isLoginRequired() (bool, error) {
//	if r.definition.Login.IsEmpty() {
//		return false, nil
//	} else if r.definition.Login.Test.IsEmpty() {
//		return true, nil
//	}
//	isLoggedIn := r.state.GetBool("loggedIn")
//	if !isLoggedIn {
//		return true, nil
//	}
//	r.logger.Debug("Testing if login is needed")
//	// HealthCheck if the login page is valid
//	match, err := r.matchPageTestBlock(r.definition.Login.Test)
//	if err != nil {
//		return true, err
//	}
//
//	if match {
//		r.logger.Debug("No login needed, already logged in")
//		return false, nil
//	}
//
//	r.logger.Debug("Login is required")
//	return true, nil
//}

// extractLoginInput gets the configured input fields and vals for the login.
func (l *BrowsingSession) extractLoginInput() (map[string]string, error) {
	result := map[string]string{}
	loginUrl, _ := l.urlContext.GetFullURL(l.loginBlock.Path)

	// Get configuration for the Indexer so we can login
	cfg, err := l.config.GetSite(l.site)
	if err != nil {
		return nil, err
	}

	ctx := struct {
		Config map[string]string
	}{
		cfg,
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

	return l.contentFetcher.FetchURL(initURL)
}

func (l *BrowsingSession) login() error {
	//if l.browser == nil {
	//	r.createBrowser()
	//	if !r.keepSessions {
	//		defer r.releaseBrowser()
	//	}
	//}
	//filterLogger = r.logger
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
		if err = l.loginBlock.hasError(l.browser); err != nil {
			l.logger.WithError(err).Error("Failed to login")
			return &LoginError{err}
		}
	}
	// HealthCheck if the login went ok
	match, err := l.verifyLogin()
	if err != nil {
		return err
	} else if !match {
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

	l.browser.SetCookieJar(cj)
	return nil
}

func (l *BrowsingSession) loginViaForm(loginURL, formSelector string, vals map[string]string) error {
	if err := l.contentFetcher.FetchURL(loginURL); err != nil {
		return err
	}

	fm, err := l.browser.Form(formSelector)
	if err != nil {
		return err
	}

	for name, value := range vals {
		if err = fm.Input(name, value); err != nil {
			//r.logger.WithError(err).Error("Filling input failed")
			return err
		}
	}

	// Maybe we don't need to cache the current browser page
	// defer r.cachePage()
	if err = fm.Submit(); err != nil {
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

	return l.contentFetcher.Post(loginURL, data, false)
}

func (l *BrowsingSession) setup() error {
	if !l.isRequired() {
		return nil
	}
	if err := l.login(); err != nil {
		l.logger.WithError(err).Error("Login failed")
		//l.noteError(err)
		return err
	}
	match, err := l.verifyLogin()
	if err != nil {
		return err
	}

	if match {
		l.state = LoggedIn
	} else {
		return fmt.Errorf("couldn't match login selector")
	}
	return nil
}
