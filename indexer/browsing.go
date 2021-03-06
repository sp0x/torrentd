package indexer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/f2prateek/train"
	trainlog "github.com/f2prateek/train/log"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf"
	"github.com/sp0x/surf/browser"
	"github.com/spf13/viper"
	"go.zoe.im/surferua"
	"golang.org/x/net/proxy"

	"github.com/sp0x/torrentd/indexer/source"
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

func createContentFetcher(r *Runner) *source.WebClient {
	browsr := surf.NewBrowser()
	userAgent := surferua.New().Desktop().Chrome().String()
	browsr.SetUserAgent(userAgent)
	browsr.SetEncoding(r.definition.Encoding)
	browsr.SetAttribute(browser.SendReferer, true)
	browsr.SetAttribute(browser.MetaRefreshHandling, true)
	browsr.SetRateLimit(r.definition.RateLimit)

	transport, err := r.createTransport()
	if err != nil {
		panic(err)
	}

	if r.options.Transport != nil {
		transport = r.options.Transport
	}

	switch os.Getenv("DEBUG_HTTP") {
	case "1", "true", "basic":
		browsr.SetTransport(train.TransportWith(transport, trainlog.New(os.Stderr, trainlog.Basic)))
	case "body":
		browsr.SetTransport(train.TransportWith(transport, trainlog.New(os.Stderr, trainlog.Body)))
	case "":
		browsr.SetTransport(transport)
	default:
		panic("Unknown value for DEBUG_HTTP")
	}
	fetchOptions := source.FetchOptions{
		ShouldDumpData: viper.GetBool("dump"),
		FakeReferer:    true,
		UserAgent:      userAgent,
	}
	contentFetcher := source.NewWebContentFetcher(browsr, r, fetchOptions)
	return contentFetcher
}
