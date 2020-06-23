package indexer

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/sp0x/surf/jar"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache/mocks"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/source"
	mocks2 "github.com/sp0x/torrentd/indexer/source/mocks"
	"github.com/sp0x/torrentd/storage/indexing"
	"github.com/sp0x/torrentd/torznab"
	"strings"
	"testing"
)

func TestRunner_Search(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	connectivityTester := mocks.NewMockConnectivityTester(ctrl)
	contentFetcher := mocks2.NewMockContentFetcher(ctrl)
	siteUrl := "http://localhost"
	//The browser should be set
	connectivityTester.EXPECT().SetBrowser(gomock.Any()).AnyTimes()
	//The correct url should be tested
	connectivityTester.EXPECT().IsOkAndSet(siteUrl, gomock.Any()).
		Return(true).AnyTimes()

	index := &IndexerDefinition{Site: "zamunda.net", Name: "zamunda"}
	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	runner := NewRunner(index, RunnerOpts{
		Config:     cfg,
		CachePages: false,
		Transport:  nil,
	})
	//Patch with our mocks
	runner.connectivityTester = connectivityTester
	runner.contentFetcher = contentFetcher
	//In order to use our custom content fetcher.
	runner.keepSessions = true

	query := &torznab.Query{}
	//Shouldn't be able to search with an index that has no urls
	_, err := runner.Search(query, nil)
	g.Expect(err).ToNot(gomega.BeNil())

	//Shouldn't be able to search with an index that has no search urls
	index.Links = []string{siteUrl}
	_, err = runner.Search(query, nil)
	g.Expect(err).ToNot(gomega.BeNil())

	//Shouldn't be able to search with an index that has no search urls
	index.Links = []string{siteUrl}
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
	//We need to mock our content fetching also
	contentFetcher.EXPECT().Fetch(gomock.Any()).
		Return(nil).
		Do(func(target *source.SearchTarget) {
			dom, _ := goquery.NewDocumentFromReader(strings.NewReader(`
<div>b<div class="a">d<a href="/lol">sd</a></div></div>`))
			fakeState := &jar.State{Dom: dom}
			runner.browser.SetState(fakeState)
		})
	srch, err := runner.Search(query, nil)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(srch).ToNot(gomega.BeNil())
	g.Expect(len(srch.GetResults())).To(gomega.Equal(1))
	firstDoc := srch.GetResults()[0]
	g.Expect(firstDoc.GUID != "").To(gomega.BeTrue())
	var foundDoc search.ExternalResultItem
	guidQuery := indexing.NewQuery()
	guidQuery.Put("GUID", firstDoc.GUID)
	g.Expect(runner.Storage.Find(guidQuery, &foundDoc)).To(gomega.BeNil())

}
