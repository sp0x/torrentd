package indexer

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/url"
)

//extractInputLogins gets the configured input fields and vals for the login.
func (r *Runner) extractInputLogins() (map[string]string, error) {
	result := map[string]string{}
	//Get configuration for the Indexer so we can login
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

func (r *Runner) login() error {
	if r.browser == nil {
		r.createBrowser()
		if !r.keepSessions {
			defer r.releaseBrowser()
		}
	}
	filterLogger = r.logger
	loginUrl, err := r.resolveIndexerPath(r.definition.Login.Path)
	if err != nil {
		return err
	}

	loginValues, err := r.extractInputLogins()
	if err != nil {
		return err
	}
	if loginValues["login_username"] == "<no value>" && loginValues["login_password"] == "<no value>" {
		return &LoginError{errors.New("no login details configured")}
	}
	switch r.definition.Login.Method {
	case "", loginMethodForm:
		if err = r.loginViaForm(loginUrl, r.definition.Login.FormSelector, loginValues); err != nil {
			return err
		}
	case loginMethodPost:
		if err = r.loginViaPost(loginUrl, loginValues); err != nil {
			return err
		}
	case loginMethodCookie:
		if err = r.loginViaCookie(loginUrl, loginValues["cookie"]); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown login method %q", r.definition.Login.Method)
	}
	// Get the error
	if len(r.definition.Login.Error) > 0 {
		if err = r.definition.Login.hasError(r.browser); err != nil {
			r.logger.WithError(err).Error("Failed to login")
			return &LoginError{err}
		}
	}
	//Check if the login went ok
	match, err := r.matchPageTestBlock(r.definition.Login.Test)
	if err != nil {
		return err
	} else if !match {
		return errors.New("login check after login failed. no matches found")
	}

	r.logger.Debug("Successfully logged in")
	r.state.Set("loggedIn", true)
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
	//Maybe we don't need to cache the current brower page
	//defer r.cachePage()
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

	return r.postToPage(loginURL, data, false)
}
