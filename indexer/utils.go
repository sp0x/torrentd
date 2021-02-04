package indexer

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func firstString(obj interface{}) string {
	switch typedObj := obj.(type) {
	case string:
		return typedObj
	case []string:
		if len(typedObj) == 0 {
			return typedObj[0]
		}
		return ""
	default:
		return fmt.Sprintf("%v", obj)
	}
}

// URLContext stores a working base URL that can be used for relative lookups
type URLContext struct {
	baseURL *url.URL
}

func (u *URLContext) GetFullURL(partialURL string) (string, error) {
	if strings.HasPrefix(partialURL, "magnet:") {
		return partialURL, nil
	}

	// Get the baseURL url of the Indexer
	if u.baseURL == nil {
		return "", errors.New("base url is nil")
	}

	partialURLParsed, err := url.Parse(partialURL)
	if err != nil {
		return "", err
	}
	// Resolve the url
	resolved := u.baseURL.ResolveReference(partialURLParsed)
	return resolved.String(), nil
}

func (r *Runner) GetURLContext() (*URLContext, error) {
	urlc := &URLContext{}
	if u := r.browser.Url(); u != nil {
		urlc.baseURL = u
		return urlc, nil
	}
	configURL, ok, _ := r.options.Config.GetSiteOption(r.definition.Site, "url")
	if ok && r.testURL(configURL) {
		resolved, _ := url.Parse(configURL)
		urlc.baseURL = resolved
		return urlc, nil
	}

	for _, u := range r.definition.Links {
		if u != configURL && r.testURL(u) {
			resolved, err := url.Parse(u)
			if err != nil {
				continue
			}
			urlc.baseURL = resolved
			return urlc, nil
		}
	}
	return nil, errors.New("no working urls found")
}

func parseCookieString(cookie string) []*http.Cookie {
	h := http.Header{"Cookie": []string{cookie}}
	r := http.Request{Header: h}
	return r.Cookies()
}