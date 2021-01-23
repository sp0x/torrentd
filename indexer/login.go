package indexer

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"
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

type LoginContext struct {
	definition loginBlock
	test       pageTestBlock
	state      LoginState
}

func newLoginContext(loginBlock loginBlock, testBlock pageTestBlock) *LoginContext {
	lc := &LoginContext{}
	lc.definition = loginBlock
	lc.test = testBlock
	if loginBlock.IsEmpty() {
		lc.state = NoLoginRequired
	} else {
		lc.state = LoginRequired
	}
	return lc
}

func (l LoginContext) isRequired() bool {
	return l.state == LoginRequired ||
		l.state == LoginFailed ||
		l.state == LoginExpired
	// if l.definition.IsEmpty() {
	//	return false
	//} else if l.test.IsEmpty() {
	//	// No way to test if we're logged in
	//	return true
	//}
}

func (l *LoginContext) test() {
	// r.logger.Debug("Testing if login is needed")
	// HealthCheck if the login page is valid
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

// isLoginRequired Checks if login is required for the given Indexer
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
	// HealthCheck if the login page is valid
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

// extractInputLogins gets the configured input fields and vals for the login.
func (r *Runner) extractInputLogins() (map[string]string, error) {
	result := map[string]string{}
	// Get configuration for the Indexer so we can login
	cfg, err := r.opts.Config.GetSite(r.definition.Name)
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

		if val == "{{ .Config.password }}" && resolved == emptyValue {
			return nil, fmt.Errorf("no password was configured for input `%s` @%s", name, r.definition.Site)
		}
		result[name] = resolved
	}

	return result, nil
}

func (r *Runner) login() error {
	if r.browser == nil {
		r.createBrowser()
		if !r.keepSessions {
			defer r.releaseBrowser()
		}
	}
	filterLogger = r.logger
	loginURL, err := r.getFullURLInIndex(r.definition.Login.Path)
	if err != nil {
		return err
	}

	loginValues, err := r.extractInputLogins()
	if err != nil {
		return err
	}

	err = r.initLogin()
	if err != nil {
		return err
	}

	switch r.definition.Login.Method {
	case "", loginMethodForm:
		if err = r.loginViaForm(loginURL, r.definition.Login.FormSelector, loginValues); err != nil {
			return err
		}
	case loginMethodPost:
		if err = r.loginViaPost(loginURL, loginValues); err != nil {
			return err
		}
	case loginMethodCookie:
		cookieVal := loginValues["cookie"]
		if cookieVal == emptyValue {
			return &LoginError{errors.New("no login cookie configured")}
		}
		if err = r.loginViaCookie(loginURL, loginValues["cookie"]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown login method %q for site %s", r.definition.Login.Method, r.Site())
	}
	// Get the error
	if len(r.definition.Login.Error) > 0 {
		if err = r.definition.Login.hasError(r.browser); err != nil {
			r.logger.WithError(err).Error("Failed to login")
			return &LoginError{err}
		}
	}
	// HealthCheck if the login went ok
	match, err := r.matchPageTestBlock(r.definition.Login.Test)
	if err != nil {
		return err
	} else if !match {
		hasPass := loginValues["login_password"] != emptyValue
		if _, ok := loginValues["login_password"]; !ok {
			hasPass = false
		}
		return fmt.Errorf("login check after login failed. no matches found. user: %s, using pass: %v", loginValues["login_username"], hasPass)
	}

	r.logger.Debug("Successfully logged in")
	r.state.Set("loggedIn", true)
	return nil
}

func (r *Runner) loginViaForm(loginURL, formSelector string, vals map[string]string) error {
	r.logger.
		WithFields(logrus.Fields{"url": loginURL, "form": formSelector, "vals": vals}).
		Debugf("Filling and submitting login form")

	if err := r.contentFetcher.FetchURL(loginURL); err != nil {
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
	// Maybe we don't need to cache the current brower page
	// defer r.cachePage()
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

	return r.contentFetcher.Post(loginURL, data, false)
}
