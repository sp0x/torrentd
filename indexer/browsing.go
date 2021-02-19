package indexer

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/sp0x/torrentd/indexer/source"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/f2prateek/train"
	trainlog "github.com/f2prateek/train/log"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf"
	"github.com/sp0x/surf/agent"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"
	"github.com/spf13/viper"
	"golang.org/x/net/proxy"

	"github.com/sp0x/torrentd/indexer/source/web"
)

func (r *Runner) createTransport() (http.RoundTripper, error) {
	var t http.Transport
	var custom bool
	// If we have a proxy to use
	if proxyAddr, isset := os.LookupEnv("SOCKS_PROXY"); isset {
		r.logger.
			WithFields(logrus.Fields{"addr": proxyAddr}).
			Debugf("Using SOCKS5 proxy")

		dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("can't connect to the proxy %s: %v", proxyAddr, err)
		}

		dc := dialer.(interface {
			DialContext(ctx context.Context, network, addr string) (net.Conn, error)
		})

		t.DialContext = dc.DialContext
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

func createContentFetcher(r *Runner) source.ContentFetcher {
	//if r.keepSessions {
	//	// No need to recreate browsers if we're keeping the session
	//	if r.browser != nil {
	//		return nil
	//	}
	//}
	r.browserLock.Lock()

	if r.cookies == nil {
		r.cookies = jar.NewMemoryCookies()
	}

	bow := surf.NewBrowser()
	bow.SetUserAgent(agent.Firefox())
	bow.SetEncoding(r.definition.Encoding)
	bow.SetAttribute(browser.SendReferer, true)
	bow.SetAttribute(browser.MetaRefreshHandling, true)
	bow.SetCookieJar(r.cookies)
	bow.SetRateLimit(r.definition.RateLimit)

	transport, err := r.createTransport()
	if err != nil {
		panic(err)
	}

	if r.options.Transport != nil {
		transport = r.options.Transport
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
	fetchOptions := source.FetchOptions{
		ShouldDumpData: viper.GetBool("dump"),
		FakeReferer:    true,
	}
	//r.connectivityTester.SetBrowser(bow)
	contentFetcher := web.NewWebContentFetcher(bow, r, r.connectivityTester, fetchOptions)
	//r.browser = bow
	return contentFetcher
}

//func (r *Runner) releaseBrowser() {
//	//r.browser = nil
//	if r.contentFetcher != nil {
//		r.contentFetcher.Cleanup()
//	}
//	r.contentFetcher = nil
//	//r.connectivityTester.ClearBrowser()
//	//r.browserLock.Unlock()
//}
