package rss

import (
	"github.com/golang/mock/gomock"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/server/http/mocks"
	"testing"
)

func TestSendRssFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	context := mocks.NewMockContext(ctrl)
	var items []*search.TorrentResultItem
	items = append(items, &search.TorrentResultItem{Title: "A"})
	items = append(items, &search.TorrentResultItem{Title: "B"})
	context.EXPECT().Header("Content-Type", "application/xml;")
	context.EXPECT().String(200, gomock.Any()).
		AnyTimes()
	SendRssFeed("host.host", "namex", items, context)
}
