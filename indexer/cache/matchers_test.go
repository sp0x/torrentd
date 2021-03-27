package cache

import (
	"fmt"
	"strings"

	"github.com/golang/mock/gomock"

	"github.com/sp0x/torrentd/indexer/source"
)

type (
	ofRequest struct {
		method string
		url    string
	}
)

func OfRequest(method string, url string) gomock.Matcher {
	return &ofRequest{method, url}
}

func (o *ofRequest) Matches(x interface{}) bool {
	req := x.(*source.RequestOptions)
	testURL := req.URL.String()
	return testURL == o.url && strings.EqualFold(req.Method, o.method)
}

func (o *ofRequest) String() string {
	return fmt.Sprintf("%s: %s", o.method, o.url)
}
