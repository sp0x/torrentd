package cache

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/golang/mock/gomock"

	"github.com/sp0x/torrentd/indexer/source"
)

type (
	ofURL     struct{ t string }
	ofRequest struct {
		method string
		url    string
	}
)

func OfRequest(method string, url string) gomock.Matcher {
	return &ofRequest{method, url}
}

func OfURL(t string) gomock.Matcher {
	return &ofURL{t}
}

func (o *ofURL) Matches(x interface{}) bool {
	return x.(*url.URL).String() == o.t
}

func (o *ofRequest) Matches(x interface{}) bool {
	req := x.(*source.RequestOptions)
	testURL := req.URL.String()
	return testURL == o.url && strings.EqualFold(req.Method, o.method)
}

func (o *ofRequest) String() string {
	return fmt.Sprintf("%s: %s", o.method, o.url)
}

func (o *ofURL) String() string {
	return o.t
}
