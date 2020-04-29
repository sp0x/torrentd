package indexer

import (
	"crypto/tls"
	"fmt"
	"github.com/f2prateek/train"
	trainlog "github.com/f2prateek/train/log"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/jar"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"time"

	"github.com/sp0x/surf"
	"github.com/sp0x/surf/agent"
	"github.com/sp0x/surf/browser"
	"os"
)

func (r *Runner) createTransport() (http.RoundTripper, error) {
	var t http.Transport
	var custom bool
	//If we have a proxy to use
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
	if r.keepSessions {
		//No need to recreate browsers if we're keeping the session
		if r.browser != nil {
			return
		}
	}
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
	r.connectivityCache.SetBrowser(bow)
	r.browser = bow

	//	minDiff := 1.0 / maxPerSecond
	//	timeElapsed := time.Now().Sub(r.lastRequest)
	//	if int(timeElapsed.Seconds()) < int(minDiff) {
	//		t := r.lastRequest.Add(time.Second * time.Duration(minDiff)).Sub(time.Now())
	//		time.Sleep(t)
	//	}
}

func (r *Runner) releaseBrowser() {
	r.browser = nil
	r.connectivityCache.ClearBrowser()
	r.browserLock.Unlock()
}

//Open a desired url
func (r *Runner) openPage(u string) error {
	r.logger.WithField("url", u).
		Debug("Opening page")
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
