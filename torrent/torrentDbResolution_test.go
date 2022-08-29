package torrent

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer"
)

func TestResolveTorrents(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewGomegaWithT(t)
	defer ctrl.Finish()
	index := indexer.NewMockIndexer(ctrl)
	indexInfo := indexer.NewMockInfo(ctrl)
	index.EXPECT().HealthCheck().Return(nil)
	index.EXPECT().Info().Return(indexInfo)
	indexInfo.EXPECT().GetID().Return("IndexID1")
	cfg := &config.ViperConfig{}

	results := ResolveTorrents(indexer.IndexCollection{index}, cfg)

	g.Expect(len(results)).To(gomega.BeEquivalentTo(0))
}
