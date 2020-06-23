package indexer

import (
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/cache/mocks"
	"github.com/sp0x/torrentd/torznab"
	"testing"
)

func TestRunner_Search(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	connectivityTester := mocks.NewMockConnectivityTester(ctrl)
	siteUrl := "http://localhost"
	//The browser should be set
	connectivityTester.EXPECT().SetBrowser(gomock.Any()).AnyTimes()
	//The correct url should be tested
	connectivityTester.EXPECT().IsOkAndSet(siteUrl, gomock.Any()).
		Return(true).AnyTimes()
	//nnectivityTester.EXPECT().test

	index := &IndexerDefinition{Site: "zamunda.net", Name: "zamunda"}
	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	runner := NewRunner(index, RunnerOpts{
		Config:     cfg,
		CachePages: false,
		Transport:  nil,
	})
	runner.connectivityTester = connectivityTester

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
	}
	srch, err := runner.Search(query, nil)
	g.Expect(srch).ToNot(gomega.BeNil())
}
