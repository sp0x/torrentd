package indexer

import (
	"bytes"
	"fmt"
	//"bytes"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"io"
)

func (r *Runner) downloadsNeedResolution() bool {
	if _, ok := r.failingSearchFields["download"]; ok {
		return true
	}
	return false
}

func (r *Runner) Open(scrapeResultItem search.ResultItemBase) (*ResponseProxy, error) {
	r.createBrowser()
	if required, err := r.isLoginRequired(); required {
		if err := r.login(); err != nil {
			r.logger.WithError(err).Error("Login failed")
			return nil, nil
		}
	} else if err != nil {
		return nil, nil
	}
	scrapeItem := scrapeResultItem.AsScrapeItem()
	sourceLink := scrapeItem.SourceLink
	//If the download needs to be resolved
	if scrapeItem.SourceLink == "" || r.downloadsNeedResolution() {
		//Resolve the url
		downloadItem := r.failingSearchFields["download"]
		err := r.contentFetcher.FetchUrl(scrapeItem.Link)
		if err != nil {
			return nil, err
		}
		scrapeItem := &DomScrapeItem{r.browser.Dom()}
		downloadLink, err := r.extractField(scrapeItem, &downloadItem)
		if err != nil {
			return nil, nil
		}
		sourceLink = firstString(downloadLink)
	}
	fullUrl, err := r.getFullUrlInIndex(sourceLink)
	if err != nil {
		return nil, err
	}
	browserClone := r.browser.NewTab()
	browserClone.SetEncoding("")
	if err := browserClone.Open(fullUrl); err != nil {
		return nil, err
	}

	pipeR, pipeW := io.Pipe()
	responsePx := &ResponseProxy{
		Reader:            pipeR,
		ContentLengthChan: make(chan int64),
	}
	//Start a goroutine and write the response of the download to the pipe
	go func() {
		defer func() {
			errx := pipeW.Close()
			if errx != nil {
				fmt.Printf("%v", errx)
			}
		}()
		if !r.keepSessions {
			defer r.releaseBrowser()
		}
		downloadBuffer := bytes.NewBuffer([]byte{})
		n, err := browserClone.Download(downloadBuffer)
		if err != nil {
			r.logger.Errorf("Error downloading: %v", err)
		} else {
			responsePx.ContentLengthChan <- n
			_, err = io.Copy(pipeW, downloadBuffer)
			if err != nil {
				r.logger.Errorf("Error piping download: %v", err)
				return
			}
			r.logger.WithFields(logrus.Fields{"url": fullUrl}).
				Infof("Downloaded %d bytes", n)
		}
	}()
	return responsePx, nil
}

func (r *Runner) Download(u string) (*ResponseProxy, error) {
	srcItem := search.ScrapeResultItem{}
	srcItem.SourceLink = u
	return r.Open(&srcItem)
}
