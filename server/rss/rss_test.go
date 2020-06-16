package rss

import (
	"github.com/golang/mock/gomock"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/server/rss/mocks"
	"testing"
)

func TestSendRssFeed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	context := mocks.NewMockHttpContext(ctrl)
	var items []search.ExternalResultItem
	items = append(items, search.ExternalResultItem{ResultItem: search.ResultItem{Title: "A"}})
	items = append(items, search.ExternalResultItem{ResultItem: search.ResultItem{Title: "B"}})
	context.EXPECT().Header("Content-Type", "application/xml;")
	context.EXPECT().String(200, gomock.Any()).
		AnyTimes()
	SendRssFeed("host.host", "namex", items, context)
}
