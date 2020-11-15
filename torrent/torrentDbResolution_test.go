package torrent

import (
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega"
	"github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/mocks"
	"testing"
)

func TestResolveTorrents(t *testing.T) {
	ctrl := gomock.NewController(t)
	g := gomega.NewGomegaWithT(t)
	defer ctrl.Finish()

	index := mocks.NewMockIndexer(ctrl)
	index.EXPECT().Check().Return(nil)
	cfg := &config.ViperConfig{}
	results := ResolveTorrents(index, cfg)

	g.Expect(len(results)).To(gomega.BeEquivalentTo(0))
}
