package indexer

import (
	"bytes"
	"fmt"
	"io"
	"net/url"

	"github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/source"
)

func (r *Runner) downloadsNeedResolution() bool {
	if _, ok := r.failingSearchFields["download"]; ok {
		return true
	}
	return false
}

func (r *Runner) Open(scrapeResultItem search.ResultItemBase) (*ResponseProxy, error) {
	_, err := r.sessions.acquire()
	if err != nil {
		return nil, err
	}
	scrapeItem := scrapeResultItem.AsScrapeItem()
	sourceLink := scrapeItem.SourceLink
	// If the download needs to be resolved
	if scrapeItem.SourceLink == "" || r.downloadsNeedResolution() {
		// Resolve the url
		downloadItem := r.failingSearchFields["download"]
		scrapeLink, err := url.Parse(scrapeItem.Link)
		if err != nil {
			return nil, err
		}
		result, err := r.contentFetcher.Fetch(source.NewRequestOptions(scrapeLink))
		if err != nil {
			return nil, err
		}
		if html, ok := result.(*source.HTMLFetchResult); ok {
			scrapeItem := source.NewDOMScrapeItem(html.DOM)
			downloadLink, err := r.extractField(scrapeItem, &downloadItem)
			if err != nil {
				return nil, nil
			}
			sourceLink = firstString(downloadLink)
		}
	}
	fullURL, err := r.urlResolver.Resolve(sourceLink)
	if err != nil {
		return nil, err
	}

	cf := r.contentFetcher.Clone()
	if _, err := cf.Open(&source.RequestOptions{
		NoEncoding: true,
		URL:        fullURL,
		Method:     "get",
	}); err != nil {
		return nil, err
	}

	responsePx, pipeW := NewResponseProxy()

	// Start a goroutine and write the response of the download to the pipe
	go func() {
		defer func() {
			errx := pipeW.Close()
			if errx != nil {
				fmt.Printf("%v", errx)
			}
		}()

		downloadBuffer := bytes.NewBuffer([]byte{})
		n, err := cf.Download(downloadBuffer)
		if err != nil {
			r.logger.Errorf("Error downloading: %v", err)
		} else {
			responsePx.ContentLengthChan <- n
			_, err = io.Copy(pipeW, downloadBuffer)
			if err != nil {
				r.logger.Errorf("Error piping download: %v", err)
				return
			}
			r.logger.WithFields(logrus.Fields{"url": fullURL}).
				Infof("Downloaded %d bytes", n)
		}
	}()
	return responsePx, nil
}

func (r *Runner) Download(url string) (*ResponseProxy, error) {
	srcItem := search.ScrapeResultItem{}
	srcItem.SourceLink = url
	return r.Open(&srcItem)
}
