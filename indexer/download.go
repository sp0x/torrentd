package indexer

import (
	"github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer/search"
	"io"
)

func (r *Runner) downloadsNeedResolution() bool {
	if _, ok := r.failingSearchFields["download"]; ok {
		return true
	}
	return false
}

func (r *Runner) Open(s *search.ExternalResultItem) (io.ReadCloser, error) {
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
		r.openPage(s.Link)
		downloadLink, err := r.extractField(r.browser.Dom(), downloadItem)
		if err != nil {
			return nil, nil
		}
		sourceLink = downloadLink
	}
	fullUrl, err := r.resolveIndexerPath(sourceLink)
	if err != nil {
		return nil, err
	}

	if err := r.browser.Open(fullUrl); err != nil {
		return nil, err
	}
	pipeR, pipeW := io.Pipe()
	go func() {
		defer pipeW.Close()
		if !r.keepSessions {
			defer r.releaseBrowser()
		}
		n, err := r.browser.Download(pipeW)
		if err != nil {
			r.logger.Error(err)
		}
		r.logger.WithFields(logrus.Fields{"url": fullUrl}).Debugf("Downloaded %d bytes", n)
	}()

	return pipeR, nil
}

func (r *Runner) Download(u string) (io.ReadCloser, error) {
	srcItem := search.ExternalResultItem{}
	srcItem.SourceLink = u
	return r.Open(&srcItem)
	//r.createBrowser()
	//
	//if required, err := r.isLoginRequired(); required {
	//	if err := r.login(); err != nil {
	//		r.logger.WithError(err).Error("Login failed")
	//		return nil, http.Header{}, err
	//	}
	//} else if err != nil {
	//	return nil, http.Header{}, err
	//}
	////If the download needs to be resolved
	//if r.downloadsNeedResolution() {
	//	//Resolve the url
	//	downloadItem := r.failingSearchFields["download"]
	//	r.openPage(u)
	//	downloadLink, err := r.extractField(r.browser.Dom(), downloadItem)
	//	if err != nil {
	//		return nil, nil, err
	//	}
	//	u = downloadLink
	//}
	//
	//fullUrl, err := r.resolveIndexerPath(u)
	//if err != nil {
	//	return nil, http.Header{}, err
	//}
	//
	//if err := r.browser.Open(fullUrl); err != nil {
	//	return nil, http.Header{}, err
	//}
	//
	//pipeR, pipeW := io.Pipe()
	//go func() {
	//	defer pipeW.Close()
	//	if !r.keepSessions {
	//		defer r.releaseBrowser()
	//	}
	//	n, err := r.browser.Download(pipeW)
	//	if err != nil {
	//		r.logger.Error(err)
	//	}
	//	r.logger.WithFields(logrus.Fields{"url": fullUrl}).Debugf("Downloaded %d bytes", n)
	//}()
	//
	//return pipeR, r.browser.ResponseHeaders(), nil
}
