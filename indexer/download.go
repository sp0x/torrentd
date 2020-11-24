package indexer

import (
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

func (r *Runner) Open(s *search.ExternalResultItem) (*ResponseProxy, error) {
	r.createBrowser()
	if required, err := r.isLoginRequired(); required {
		if err := r.login(); err != nil {
			r.logger.WithError(err).Error("Login failed")
			return nil, nil
		}
	} else if err != nil {
		return nil, nil
	}
	sourceLink := s.SourceLink
	//If the download needs to be resolved
	if s.SourceLink == "" || r.downloadsNeedResolution() {
		//Resolve the url
		downloadItem := r.failingSearchFields["download"]
		err := r.contentFetcher.FetchUrl(s.Link)
		if err != nil {
			return nil, err
		}
		downloadLink, err := r.extractField(r.browser.Dom(), &downloadItem)
		if err != nil {
			return nil, nil
		}
		sourceLink = firstString(downloadLink)
	}
	fullUrl, err := r.resolveIndexerPath(sourceLink)
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
			_ = pipeW.Close()
		}()
		if !r.keepSessions {
			defer r.releaseBrowser()
		}
		n, err := browserClone.Download(pipeW)
		if err != nil {
			r.logger.Errorf("Error downloading: %v", err)
		} else {
			responsePx.ContentLengthChan <- n
			r.logger.WithFields(logrus.Fields{"url": fullUrl}).
				Infof("Downloaded %d bytes", n)
		}
	}()
	return responsePx, nil
}

func (r *Runner) Download(u string) (*ResponseProxy, error) {
	srcItem := search.ExternalResultItem{}
	srcItem.SourceLink = u
	return r.Open(&srcItem)
}
