package indexer

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/surf/jar"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/cache/mocks"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/source"
	mocks2 "github.com/sp0x/torrentd/indexer/source/mocks"
	"github.com/sp0x/torrentd/storage/indexing"
)

var (
	runnerSiteUrl = "http://localhost/"
	index         = &IndexerDefinition{
		Site:  "zamunda.net",
		Name:  "zamunda",
		Links: []string{runnerSiteUrl},
	}
)

//
//func TestRunner_SearchShouldWorkWithAnOptimisticCache(t *testing.T) {
//	g := gomega.NewWithT(t)
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//	mockedBrowser := &mocks.MockedBrowser{CanOpen: false}
//	cfg := &config.ViperConfig{}
//	_ = cfg.Set("db", tempfile())
//	_ = cfg.Set("storage", "boltdb")
//	tmpIndex := *index
//	index := &tmpIndex
//	index.Search = searchBlock{
//		Path: "/",
//		Rows: rowsBlock{
//			selectorBlock: selectorBlock{
//				Selector: "div.a",
//			},
//		},
//		Fields: fieldsListBlock{fieldBlock{
//			Field: "fieldA",
//			Block: selectorBlock{
//				Selector: "a",
//			},
//		}},
//	}
//
//	runner := NewRunner(index, RunnerOpts{
//		Config:     cfg,
//		CachePages: false,
//		Transport:  nil,
//	})
//	//If we use a normal connectivity cache, this test would fail, because it validates the connectivity early.
//	//conCache, _ := cache.NewConnectivityCache()
//	//runner.connectivityTester = conCache
//	//Patch with our mocks
//	//runner.connectivityTester = connectivityTester
//	runner.createBrowser()
//	runner.connectivityTester.SetBrowser(mockedBrowser)
//	//runner.contentFetcher = contentFetcher
//	runner.browser = mockedBrowser
//
//	//We expect a single fetch, with the optimistic cache
//	//	contentFetcher.EXPECT().Fetch(gomock.Any()).
//	//		Times(1).
//	//		Return(errors.New("couldn't connect")).
//	//		Do(func(target *source.SearchTarget) {
//	//			dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
//	//<div>b<div class="a">d<a href="/lol">sd</a></div></div>
//	//<div class="b"><a>val1</a><p>parrot</p></div>`))
//	//			fakeState := &jar.State{Dom: dom}
//	//			runner.browser.SetState(fakeState)
//	//		})
//
//	dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`<div></div>`))
//	mockedBrowser.SetState(&jar.State{Dom: dom})
//	_, err := runner.Search(emptyQuery, nil)
//	g.Expect(err).ToNot(gomega.BeNil())
//	//The connectivity tester should remember that that url is bad
//	g.Expect(runner.connectivityTester.IsOk(runnerSiteUrl)).To(gomega.BeFalse())
//
//	//Try again, now with a working browser
//	contentFetcher := mocks2.NewMockContentFetcher(ctrl)
//	mockedBrowser.CanOpen = true
//	contentFetcher.EXPECT().Fetch(gomock.Any()).
//		Times(1).
//		Return(nil).
//		Do(func(target *source.SearchTarget) {
//			dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
//<div>b<div class="a">d<a href="/lol">sd</a></div></div>
//<div class="b"><a>val1</a><p>parrot</p></div>`))
//			fakeState := &jar.State{Dom: dom}
//			runner.browser.SetState(fakeState)
//		})
//	runner.contentFetcher = contentFetcher
//	_, err = runner.Search(emptyQuery, nil)
//	g.Expect(err).To(gomega.BeNil())
//	g.Expect(runner.connectivityTester.IsOk(runnerSiteUrl)).To(gomega.BeTrue())
//}

func TestRunner_Search(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	connectivityTester := mocks.NewMockConnectivityTester(ctrl)
	contentFetcher := mocks2.NewMockContentFetcher(ctrl)
	// The browser should be set
	connectivityTester.EXPECT().SetBrowser(gomock.Any()).AnyTimes()
	// The correct url should be tested
	connectivityTester.EXPECT().IsOkAndSet(runnerSiteUrl, gomock.Any()).
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
	runner.createBrowser()
	runner.contentFetcher = contentFetcher
	// In order to use our custom content fetcher.
	runner.keepSessions = true

	contentFetcher.EXPECT().Fetch(gomock.Any()).
		AnyTimes().
		Return(nil).
		Do(func(target *source.SearchTarget) {
			dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
<div>b<div class="a">d<a href="/lol">sd</a></div></div>
<div class="b"><a>val1</a><p>parrot</p></div>`))
			fakeState := &jar.State{Dom: dom}
			runner.browser.SetState(fakeState)
		})

	// Shouldn't be able to search with an index that has no urls
	_, err := runner.Search(emptyQuery, nil)
	g.Expect(err).ToNot(gomega.BeNil())

	// Shouldn't be able to search with an index that has no search urls
	index.Links = []string{runnerSiteUrl}
	_, err = runner.Search(emptyQuery, nil)
	g.Expect(err).ToNot(gomega.BeNil())

	// Shouldn't be able to search with an index that has no search urls
	index.Links = []string{runnerSiteUrl}
	index.Search = searchBlock{
		Path: "/",
		Rows: rowsBlock{
			selectorBlock: selectorBlock{
				Selector: "div.a",
			},
		},
		Fields: fieldsListBlock{fieldBlock{
			Field: "fieldA",
			Block: selectorBlock{
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
	index := &IndexerDefinition{Site: "zamunda.net", Name: "zamunda"}
	contentFetcher := mocks2.NewMockContentFetcher(ctrl)
	connectivityTester := mocks.NewMockConnectivityTester(ctrl)
	// The browser should be set
	connectivityTester.EXPECT().SetBrowser(gomock.Any()).AnyTimes()
	// The correct url should be tested
	connectivityTester.EXPECT().IsOkAndSet(runnerSiteUrl, gomock.Any()).
		Return(true).AnyTimes()

	//-------Should be able to use unique indexes
	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	_ = cfg.Set("storage", "boltdb")
	index.Name = "other"
	index.Links = []string{runnerSiteUrl}
	index.Search = searchBlock{
		Path: "/",
		Key:  []string{"fieldB"},
		Rows: rowsBlock{
			selectorBlock: selectorBlock{
				Selector: "div.b",
			},
		},
		Fields: fieldsListBlock{
			fieldBlock{
				Field: "id",
				Block: selectorBlock{
					Selector: "a",
				},
			},
			fieldBlock{
				Field: "fieldC",
				Block: selectorBlock{
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
	runner.createBrowser()
	runner.contentFetcher = contentFetcher
	contentFetcher.EXPECT().Fetch(gomock.Any()).
		Return(nil).
		Do(func(target *source.SearchTarget) {
			dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
<div>b<div class="a">d<a href="/lol">sd</a></div></div>
<div class="b"><a>val1</a><p>parrot</p></div>`))
			fakeState := &jar.State{Dom: dom}
			runner.browser.SetState(fakeState)
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
	cachedCon, _ := cache.NewConnectivityCache()
	r.connectivityTester = cachedCon
	g.Expect(r.testURLWorks("http://example.com")).To(gomega.BeFalse())
}

func TestRunner_testUrlWorks_ShouldWorkWithOptimisticCaching(t *testing.T) {
	g := gomega.NewWithT(t)
	r := Runner{}
	r.logger = log.New()
	url := "http://example.com"
	optimisticCacheCon, _ := cache.NewOptimisticConnectivityCache()
	r.connectivityTester = optimisticCacheCon
	g.Expect(r.testURLWorks(url)).To(gomega.BeTrue())
	r.connectivityTester.Invalidate(url)
}
