package indexer

import (
	"net/url"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/cache/mocks"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/source"
	mocks2 "github.com/sp0x/torrentd/indexer/source/mocks"
	"github.com/sp0x/torrentd/storage/indexing"
)

var (
	runnerSiteURL = "http://localhost/"
	index         = &Definition{
		Site:  "zamunda.net",
		Name:  "zamunda",
		Links: []string{runnerSiteURL},
	}
)

func TestRunner_Search(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	connectivityTester := mocks.NewMockConnectivityTester(ctrl)
	contentFetcher := mocks2.NewMockContentFetcher(ctrl)
	// The browser should be set
	// connectivityTester.EXPECT().SetBrowser(gomock.Any()).AnyTimes()
	// The correct url should be tested
	connectivityTester.EXPECT().IsOkAndSet(runnerSiteURL, gomock.Any()).
		Return(true).AnyTimes()

	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	_ = cfg.Set("storage", "boltdb")
	runner := NewRunner(index, RunnerOpts{
		Config:     cfg,
		CachePages: false,
		Transport:  nil,
	})
	// Patch with our mocks
	runner.connectivityTester = connectivityTester
	runner.contentFetcher = contentFetcher
	// In order to use our custom content fetcher.

	contentFetcher.EXPECT().Fetch(gomock.Any()).
		AnyTimes().
		Return(nil).
		Do(func(target *source.FetchOptions) {
			// dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
			// <div>b<div class="a">d<a href="/lol">sd</a></div></div>
			// <div class="b"><a>val1</a><p>parrot</p></div>`))
			// fakeState := &jar.State{Dom: dom}
			// runner.browser.SetState(fakeState)
		})

	// Shouldn't be able to search with an index that has no urls
	_, err := runner.Search(emptyQuery, nil)
	g.Expect(err).ToNot(gomega.BeNil())

	// Shouldn't be able to search with an index that has no search urls
	index.Links = []string{runnerSiteURL}
	_, err = runner.Search(emptyQuery, nil)
	g.Expect(err).ToNot(gomega.BeNil())

	// Shouldn't be able to search with an index that has no search urls
	index.Links = []string{runnerSiteURL}
	index.Search = searchBlock{
		Path: "/",
		Rows: rowsBlock{
			SelectorBlock: source.SelectorBlock{
				Selector: "div.a",
			},
		},
		Fields: fieldsListBlock{fieldBlock{
			Field: "fieldA",
			Block: source.SelectorBlock{
				Selector: "a",
			},
		}},
	}
	runner.contentFetcher = contentFetcher
	// We need to mock our content fetching also
	srch, err := runner.Search(emptyQuery, nil)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(srch).ToNot(gomega.BeNil())
	g.Expect(len(srch.GetResults()) > 0).To(gomega.BeTrue())
	firstDoc := srch.GetResults()[0]

	g.Expect(firstDoc.UUID() != "").To(gomega.BeTrue())
	var foundDoc search.ScrapeResultItem
	guidQuery := indexing.NewQuery()
	guidQuery.Put("UUID", firstDoc.UUID())
	storage := getIndexStorage(runner, cfg)
	g.Expect(storage.Find(guidQuery, &foundDoc)).To(gomega.BeNil())
	g.Expect(foundDoc.UUIDValue).To(gomega.Equal(firstDoc.UUID()))
	g.Expect(foundDoc.ModelData["fieldA"]).To(gomega.Equal("sd"))
	storage.Close()
}

var emptyQuery = &search.Query{}

func Test_ShouldUseUniqueIndexes(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	index := &Definition{Site: "zamunda.net", Name: "zamunda"}
	contentFetcher := mocks2.NewMockContentFetcher(ctrl)
	connectivityTester := mocks.NewMockConnectivityTester(ctrl)
	// The browser should be set
	// connectivityTester.EXPECT().SetBrowser(gomock.Any()).AnyTimes()
	// The correct url should be tested
	connectivityTester.EXPECT().IsOkAndSet(runnerSiteURL, gomock.Any()).
		Return(true).AnyTimes()

	// -------Should be able to use unique indexesCollection
	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	_ = cfg.Set("storage", "boltdb")
	index.Name = "other"
	index.Links = []string{runnerSiteURL}
	index.Search = searchBlock{
		Path: "/",
		Key:  []string{"fieldB"},
		Rows: rowsBlock{
			SelectorBlock: source.SelectorBlock{
				Selector: "div.b",
			},
		},
		Fields: fieldsListBlock{
			fieldBlock{
				Field: "id",
				Block: source.SelectorBlock{
					Selector: "a",
				},
			},
			fieldBlock{
				Field: "fieldC",
				Block: source.SelectorBlock{
					Selector: "p",
				},
			},
		},
	}

	runner := NewRunner(index, RunnerOpts{
		Config:     cfg,
		CachePages: false,
		Transport:  nil,
	})
	// Patch with our mocks
	runner.connectivityTester = connectivityTester
	runner.contentFetcher = contentFetcher
	contentFetcher.EXPECT().Fetch(gomock.Any()).
		Return(nil).
		Do(func(target *source.FetchOptions) {
			//			dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
			// <div>b<div class="a">d<a href="/lol">sd</a></div></div>
			// <div class="b"><a>val1</a><p>parrot</p></div>`))
			// fakeState := &jar.State{Dom: dom}
			// runner.browser.SetState(fakeState)
		})
	srch, err := runner.Search(emptyQuery, nil)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(srch).ToNot(gomega.BeNil())
	g.Expect(len(srch.GetResults()) == 1).To(gomega.BeTrue())
}

func TestRunner_testURLWorks_ShouldReturnFalseIfTheUrlIsDown(t *testing.T) {
	g := gomega.NewWithT(t)
	r := Runner{}
	r.logger = log.New()
	contentFetcher := &source.WebClient{}
	cachedCon, _ := cache.NewConnectivityCache(contentFetcher)
	r.connectivityTester = cachedCon
	pURL, _ := url.Parse("http://example.com")
	g.Expect(r.urlResolver.connectivity.IsOk(pURL)).To(gomega.BeFalse())
}

func TestRunner_testUrlWorks_ShouldWorkWithOptimisticCaching(t *testing.T) {
	g := gomega.NewWithT(t)
	r := Runner{}
	r.logger = log.New()
	pURL, _ := url.Parse("http://example.com")
	contentFetcher := &source.WebClient{}
	optimisticCacheCon, _ := cache.NewOptimisticConnectivityCache(contentFetcher)
	r.connectivityTester = optimisticCacheCon
	g.Expect(r.connectivityTester.IsOk(pURL)).To(gomega.BeTrue())
	r.connectivityTester.Invalidate(pURL.String())
}
