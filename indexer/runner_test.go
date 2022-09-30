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

func getSUT(ctrl *gomock.Controller) *Runner {
	connectivityTester := mocks.NewMockConnectivityTester(ctrl)
	contentFetcher := mocks2.NewMockContentFetcher(ctrl)
	urlResolver := NewMockIURLResolver(ctrl)

	cfg := &config.ViperConfig{}
	cfg.Set("db", tempfile())
	cfg.Set("storage", "boltdb")
	index := NewRunner(getIndexDefinition(), &RunnerOpts{
		Config:     cfg,
		CachePages: false,
		Transport:  nil,
	})
	// Patch with our mocks
	index.connectivityTester = connectivityTester
	index.contentFetcher = contentFetcher
	index.urlResolver = urlResolver
	index.definition.Search.Inputs = make(map[string]string)

	// The browser should be set
	// connectivityTester.EXPECT().SetBrowser(gomock.Any()).AnyTimes()
	// The correct url should be tested
	connectivityTester.EXPECT().IsValidOrSet(runnerSiteURL, gomock.Any()).
		Return(true).
		AnyTimes()

	dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
			<div>b<div class="a">d<a href="/lol">sd</a></div></div>
			<div class="b"><a>val1</a><p>parrot</p></div>`))
	fetchResult := &source.HTMLFetchResult{DOM: dom}
	contentFetcher.EXPECT().Fetch(gomock.Any()).
		AnyTimes().
		Return(fetchResult, nil)

	return index
}

func Test_Should_NotBeAbleToSearch_WithoutURLs(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	index := getSUT(ctrl)
	urlResolver := index.urlResolver.(*MockIURLResolver)

	urlResolveMock := urlResolver.EXPECT().Resolve("/")

	// Shouldn't be able to search with an index that has no urls
	iter := search.NewIterator(search.NewQuery())
	index.definition.Search.Rows = rowsBlock{SelectorBlock: source.SelectorBlock{Selector: ".a"}}
	urlResolveMock.Return(nil, errors.New("err")).Times(1)

	fields, page := iter.Next()
	_, err := index.Search(search.NewQuery(), createWorkerJob(nil, index, fields, page))

	g.Expect(err).ToNot(gomega.BeNil())
}

func Test_Given_UrlResolverCanNotResolveUrls_Search_Should_ErrorOut(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	index := getSUT(ctrl)

	// Shouldn't be able to search with an index that has no search urls
	urlResolver := index.urlResolver.(*MockIURLResolver)
	index.urlResolver = urlResolver
	urlResolver.EXPECT().Resolve("/").Return(nil, errors.New("err")).AnyTimes()
	iter := search.NewIterator(search.NewQuery())
	index.definition.Links = []string{}

	fields, page := iter.Next()
	_, err := index.Search(search.NewQuery(), createWorkerJob(nil, index, fields, page))

	g.Expect(err).ToNot(gomega.BeNil())
}

func TestRunner_Search(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	index := getSUT(ctrl)
	connectivityTester := index.connectivityTester.(*mocks.MockConnectivityTester)
	contentFetcher := index.contentFetcher.(*mocks2.MockContentFetcher)
	urlResolver := index.urlResolver.(*MockIURLResolver)
	destURL, _ := url.Parse("http://localhost/")
	urlResolver.EXPECT().Resolve("/").Return(destURL, nil).AnyTimes()

	// Shouldn't be able to search with an index that has no search urls
	index.definition.Links = []string{runnerSiteURL}
	index.contentFetcher = contentFetcher
	index.connectivityTester = connectivityTester
	// We need to mock our content fetching also
	iter := search.NewIterator(search.NewQuery())
	fields, page := iter.Next()

	results, err := index.Search(search.NewQuery(), createWorkerJob(iter, index, fields, page))
	g.Expect(err).To(gomega.BeNil())
	g.Expect(results).ToNot(gomega.BeNil())
	g.Expect(len(results) > 0).To(gomega.BeTrue())
	firstDoc := results[0]

	g.Expect(firstDoc.AsScrapeItem().ModelData["fieldA"]).To(gomega.Not(gomega.BeEmpty()))
}

func Test_ShouldUseUniqueIndexes(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// -------Should be able to use unique indexMap
	index := getSUT(ctrl)
	index.definition.Links = []string{runnerSiteURL}
	index.definition.Search = searchBlock{
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

	urlResolver := index.urlResolver.(*MockIURLResolver)

	mockURL, _ := url.Parse("http://localhost/")
	urlResolver.EXPECT().Resolve(gomock.Any()).
		Return(mockURL, nil)

	iter := search.NewIterator(search.NewQuery())
	fields, page := iter.Next()
	results, err := index.Search(search.NewQuery(), createWorkerJob(nil, index, fields, page))

	g.Expect(err).To(gomega.BeNil())
	g.Expect(results).ToNot(gomega.BeNil())
	g.Expect(len(results) == 1).To(gomega.BeTrue())
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

func getSearchTemplateDataForNextPage(iter *search.SearchStateIterator, query *search.Query) *SearchTemplateData {
	fields, page := iter.Next()
	searchTemplateData := newSearchTemplateData(query, createWorkerJob(nil, nil, fields, page), nil)

	return searchTemplateData
}

func Test_getURLValuesForSearch_Given_RangeFieldInDefinition_Should_UseItInSearch(t *testing.T) {
	// This is not supported anymore
	t.SkipNow()
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	runner := getSUT(ctrl)
	runner.definition.Search.Inputs["rangeField"] = "{{ rng \"001\" \"010\" }}"
	query := search.NewQuery()
	iter := search.NewIterator(query)
	searchTemplateData := getSearchTemplateDataForNextPage(iter, query)

	values, err := getURLValuesForSearch(&runner.definition.Search, searchTemplateData)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(values).ToNot(gomega.BeNil())
	g.Expect(values.Encode()).To(gomega.Equal("rangeField=001"))
	values, _ = getURLValuesForSearch(&runner.definition.Search, searchTemplateData)
	//goland:noinspection GoNilness
	g.Expect(values.Encode()).To(gomega.Equal("rangeField=002"))
}

func Test_getURLValuesForSearch_Given_RangeFieldInQuery_Then_URLValues_Should_ContainRange(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	runner := getSUT(ctrl)
	runner.definition.Search.Inputs["rangeField"] = ""
	query := search.NewQuery()
	query.Fields["rangeField"] = search.NewRangeField("001", "010")
	iter := search.NewIterator(query)
	searchTemplateData := getSearchTemplateDataForNextPage(iter, query)

	values, err := getURLValuesForSearch(&runner.definition.Search, searchTemplateData)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(values).ToNot(gomega.BeNil())
	g.Expect(values.Encode()).To(gomega.Equal("rangeField=001"))

	// Iterate our state and get the new values
	searchTemplateData = getSearchTemplateDataForNextPage(iter, query)
	values, _ = getURLValuesForSearch(&runner.definition.Search, searchTemplateData)

	g.Expect(values.Encode()).To(gomega.Equal("rangeField=002"))
}
