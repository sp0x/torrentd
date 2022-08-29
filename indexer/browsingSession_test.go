package indexer

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache/mocks"
	"github.com/sp0x/torrentd/indexer/source"
	mocks2 "github.com/sp0x/torrentd/indexer/source/mocks"
)

func newTestingIndex() *Runner {
	cfg := &config.ViperConfig{}
	cfg.Set("db", tempfile())
	cfg.Set("storage", "boltdb")
	indexDef := &Definition{
		Site:  "example.com",
		Name:  "example",
		Links: []string{"http://example.com/"},
		Login: loginBlock{
			Path:   "/login",
			Method: "POST",
			Inputs: map[string]string{"key": "value"},
			Test: pageTestBlock{
				Selector: ".loggedin",
			},
			Init: initBlock{},
		},
	}
	runner := NewRunner(indexDef, &RunnerOpts{
		Config:     cfg,
		CachePages: false,
		Transport:  nil,
	})
	return runner
}

func TestNewSessionMultiplexer_ShouldCreateANumberOfSessions(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	sampleIndex := newTestingIndex()

	multiplexer, err := NewSessionMultiplexer(sampleIndex, 10)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(multiplexer.sessions)).To(gomega.Equal(10))
}

func TestNewSessionMultiplexer(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mConnectivity := mocks.NewMockConnectivityTester(ctrl)
	mContentFetcher := mocks2.NewMockContentFetcher(ctrl)
	sampleIndex := newTestingIndex()
	// sampleIndex.urlResolver.connectivity = mConnectivity

	mConnectivity.EXPECT().IsValidOrSet(OfURL("http://example.com/"), gomock.Any()).
		AnyTimes().
		Return(true)

	multiplexer, err := NewSessionMultiplexer(sampleIndex, 3)
	patchSessions(multiplexer, mContentFetcher)
	g.Expect(err).To(gomega.BeNil())

	expectLogin(mContentFetcher, "post", "http://example.com/login")
	s1, err := multiplexer.acquire()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(s1.isLoggedIn())

	expectLogin(mContentFetcher, "post", "http://example.com/login")
	s2, err := multiplexer.acquire()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(s2.isLoggedIn())

	expectLogin(mContentFetcher, "post", "http://example.com/login")
	s3, err := multiplexer.acquire()
	g.Expect(err).To(gomega.BeNil())
	g.Expect(s3.isLoggedIn())

	g.Expect(s1).ToNot(gomega.Equal(s2))
	g.Expect(s2).ToNot(gomega.Equal(s3))
	g.Expect(s3).ToNot(gomega.Equal(s1))
}

func expectLogin(cf *mocks2.MockContentFetcher, method string, url string) {
	loginPage, _ := goquery.NewDocumentFromReader(strings.NewReader(`
	<div class='loggedin'></div>`))
	loginResponse := &source.HTMLFetchResult{
		HTTPResult: source.HTTPResult{
			Response:   nil,
			StatusCode: 200,
		},
		DOM: loginPage,
	}

	cf.EXPECT().
		Post(OfRequest(method, url)).
		Return(loginResponse, nil)
}

func patchSessions(multiplexer *BrowsingSessionMultiplexer, fetcher *mocks2.MockContentFetcher) {
	for _, s := range multiplexer.sessions {
		s.contentFetcher = fetcher
	}
}
