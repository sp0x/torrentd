package indexer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/search"
)

//go:generate mockgen -source utils.go -destination=utilsMock.go -package=indexer
type IURLResolver interface {
	Resolve(partialURL string) (*url.URL, error)
}

type URLResolver struct {
	urls         []*url.URL
	connectivity cache.ConnectivityTester
	logger       *log.Logger
}

func (r *URLResolver) Resolve(partialURL string) (*url.URL, error) {
	if isUnresolvable(partialURL) {
		return url.Parse(partialURL)
	}
	for _, cURL := range r.urls {
		baseURL := cURL
		if r.connectivity.IsValidOrSet(baseURL.String(), func() bool {
			return defaultURLTester(r.connectivity, baseURL, r.logger)
		}) {
			return r.resolvePartial(baseURL, partialURL)
		}
	}

	return nil, errors.New("couldn't find a working URL")
}

func isUnresolvable(partialURL string) bool {
	return strings.HasPrefix(partialURL, "magnet:")
}

func (r *URLResolver) resolvePartial(baseURL *url.URL, partialURL string) (*url.URL, error) {
	// Search the baseURL url of the Indexes
	if baseURL == nil {
		return nil, errors.New("base url is nil")
	}

	partialURLParsed, err := url.Parse(partialURL)
	if err != nil {
		return nil, err
	}
	// Resolve the url
	resolved := baseURL.ResolveReference(partialURLParsed)
	return resolved, nil
}

// The check would be performed only if the connectivity tester doesn't have an entry for that URL
func defaultURLTester(connectivity cache.ConnectivityTester, testURL *url.URL, logger *log.Logger) bool {
	logger.WithField("url", testURL).
		Info("Checking connectivity to url")
	err := connectivity.Test(testURL.String())
	if err != nil {
		logger.WithError(err).Warn("URL check failed")
		return false
	}
	return true
}

func NewURLResolver(urls []*url.URL, connectivity *cache.ConnectivityCache) IURLResolver {
	resolver := &URLResolver{
		urls:         urls,
		connectivity: connectivity,
		logger:       log.New(),
	}
	return resolver
}

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

func newURLResolverForIndex(definition *Definition, cfg config.Config, connectivity *cache.ConnectivityCache) IURLResolver {
	var urls []*url.URL
	configURL, ok, _ := cfg.GetSiteOption(definition.Site, "url")
	if ok {
		resolved, err := url.Parse(configURL)
		if err != nil {
			urls = append(urls, resolved)
		}
	}

	for _, u := range definition.Links {
		if u != configURL {
			resolved, err := url.Parse(u)
			if err != nil {
				continue
			}
			urls = append(urls, resolved)
		}
	}
	return NewURLResolver(urls, connectivity)
}

func parseCookieString(cookie string) []*http.Cookie {
	h := http.Header{"Cookie": []string{cookie}}
	r := http.Request{Header: h}
	return r.Cookies()
}

func PrintResults(items []search.ResultItemBase, writer io.Writer) {
	for _, item := range items {
		if item.IsNew() || item.IsUpdate() {
			if item.IsNew() && !item.IsUpdate() {
				serialized, _ := json.Marshal(item)
				_, _ = fmt.Fprintf(writer, "Found new result #%s:\t%s\n",
					item.UUID(), string(serialized))
			} else {
				_, _ = fmt.Fprintf(writer, "Updated result #%s:\t%s\n",
					item.UUID(), item.String())
			}
		} else {
			_, _ = fmt.Fprintf(writer, "Result #%s:\t%s\n",
				item.UUID(), item.String())
		}
	}
}
