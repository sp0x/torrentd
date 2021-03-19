package indexer

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
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

var runnerSiteURL = "http://localhost/"

func getIndexDefinition() *Definition {
	return &Definition{
		Site:  "zamunda.net",
		Name:  "zamunda",
		Links: []string{runnerSiteURL},
		Search: searchBlock{
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
		},
	}
}

func getIndex(ctrl *gomock.Controller) *Runner {
	connectivityTester := mocks.NewMockConnectivityTester(ctrl)
	contentFetcher := mocks2.NewMockContentFetcher(ctrl)
	urlResolver := NewMockIURLResolver(ctrl)

	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	_ = cfg.Set("storage", "boltdb")
	runner := NewRunner(getIndexDefinition(), RunnerOpts{
		Config:     cfg,
		CachePages: false,
		Transport:  nil,
	})
	// Patch with our mocks
	runner.connectivityTester = connectivityTester
	runner.contentFetcher = contentFetcher
	runner.urlResolver = urlResolver

	// The browser should be set
	// connectivityTester.EXPECT().SetBrowser(gomock.Any()).AnyTimes()
	// The correct url should be tested
	connectivityTester.EXPECT().IsOkAndSet(runnerSiteURL, gomock.Any()).
		Return(true).
		AnyTimes()

	dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
			<div>b<div class="a">d<a href="/lol">sd</a></div></div>
			<div class="b"><a>val1</a><p>parrot</p></div>`))
	fetchResult := &source.HTMLFetchResult{DOM: dom}
	contentFetcher.EXPECT().Fetch(gomock.Any()).
		AnyTimes().
		Return(fetchResult, nil)

	return runner
}

func Test_ShouldntBeAbleToSearchWithoutUrls(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	runner := getIndex(ctrl)
	urlResolver := runner.urlResolver.(*MockIURLResolver)

	urlResolveMock := urlResolver.EXPECT().Resolve("/")

	// Shouldn't be able to search with an index that has no urls
	srch := search.NewSearch(search.NewQuery())
	runner.definition.Search.Rows = rowsBlock{SelectorBlock: source.SelectorBlock{Selector: ".a"}}
	urlResolveMock = urlResolveMock.Return(nil, errors.New("err")).Times(1)

	_, err := runner.Search(search.NewQuery(), srch)

	g.Expect(err).ToNot(gomega.BeNil())
}

func Test_Given_UrlResolverCanNotResolveUrls_Search_Should_ErrorOut(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	runner := getIndex(ctrl)

	// Shouldn't be able to search with an index that has no search urls
	urlResolver := runner.urlResolver.(*MockIURLResolver)
	runner.urlResolver = urlResolver
	urlResolver.EXPECT().Resolve("/").Return(nil, errors.New("err")).AnyTimes()
	srch := search.NewSearch(search.NewQuery())
	runner.definition.Links = []string{}

	_, err := runner.Search(search.NewQuery(), srch)

	g.Expect(err).ToNot(gomega.BeNil())
}

func TestRunner_Search(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	runner := getIndex(ctrl)
	connectivityTester := runner.connectivityTester.(*mocks.MockConnectivityTester)
	contentFetcher := runner.contentFetcher.(*mocks2.MockContentFetcher)
	urlResolver := runner.urlResolver.(*MockIURLResolver)
	destUrl, _ := url.Parse("http://localhost/")
	urlResolver.EXPECT().Resolve("/").Return(destUrl, nil).AnyTimes()

	// Shouldn't be able to search with an index that has no search urls
	runner.definition.Links = []string{runnerSiteURL}

	runner.contentFetcher = contentFetcher
	runner.connectivityTester = connectivityTester
	// We need to mock our content fetching also
	newSearch := search.NewSearch(nil)
	srch, err := runner.Search(search.NewQuery(), newSearch)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(srch).ToNot(gomega.BeNil())
	g.Expect(len(srch.GetResults()) > 0).To(gomega.BeTrue())
	firstDoc := srch.GetResults()[0]

	g.Expect(firstDoc.UUID() != "").To(gomega.BeTrue())
}

func Test_Given_SearchFindsResults_Then_Results_Should_BeStoredInTheIndexStorage(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	runner := getIndex(ctrl)
	urlResolver := runner.urlResolver.(*MockIURLResolver)
	destUrl, _ := url.Parse("http://localhost/")
	urlResolver.EXPECT().Resolve("/").
		Return(destUrl, nil).AnyTimes()

	newSearch := search.NewSearch(nil)
	srch, err := runner.Search(search.NewQuery(), newSearch)
	g.Expect(err).To(gomega.BeNil())
	firstDoc := srch.GetResults()[0]

	guidQuery := indexing.NewQuery()
	guidQuery.Put("UUID", firstDoc.UUID())
	storage := runner.GetStorage()
	defer storage.Close()

	var foundDoc search.ScrapeResultItem
	g.Expect(storage.Find(guidQuery, &foundDoc)).To(gomega.BeNil())
	g.Expect(foundDoc.UUIDValue).To(gomega.Equal(firstDoc.UUID()))
	g.Expect(foundDoc.ModelData["fieldA"]).To(gomega.Equal("sd"))
}

func Test_ShouldUseUniqueIndexes(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// -------Should be able to use unique indexesCollection
	runner := getIndex(ctrl)
	runner.definition.Links = []string{runnerSiteURL}
	runner.definition.Search = searchBlock{
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

	urlResolver := runner.urlResolver.(*MockIURLResolver)

	mockURL, _ := url.Parse("http://localhost/")
	urlResolver.EXPECT().Resolve(gomock.Any()).
		Return(mockURL, nil)

	srch := search.NewSearch(search.NewQuery())
	srch, err := runner.Search(search.NewQuery(), srch)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(srch).ToNot(gomega.BeNil())
	g.Expect(len(srch.GetResults()) == 1).To(gomega.BeTrue())
}

func TestRunner_testURLWorks_Should_ReturnFalseIfTheUrlIsDown(t *testing.T) {
	g := gomega.NewWithT(t)
	r := Runner{}
	r.logger = log.New()
	contentFetcher := &source.WebClient{}
	cachedCon, _ := cache.NewConnectivityCache(contentFetcher)
	r.connectivityTester = cachedCon
	pURL, _ := url.Parse("http://example.com")

	g.Expect(cachedCon.IsOk(pURL.String())).To(gomega.BeFalse())
}

func TestRunner_testUrlWorks_ShouldWorkWithOptimisticCaching(t *testing.T) {
	g := gomega.NewWithT(t)
	r := Runner{}
	r.logger = log.New()
	pURL, _ := url.Parse("http://example.com")
	contentFetcher := &source.WebClient{}
	optimisticCacheCon, _ := cache.NewOptimisticConnectivityCache(contentFetcher)
	r.connectivityTester = optimisticCacheCon

	g.Expect(r.connectivityTester.IsOk(pURL.String())).To(gomega.BeTrue())

	r.connectivityTester.Invalidate(pURL.String())
}
