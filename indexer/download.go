package indexer

import (
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func (r *Runner) downloadsNeedResolution() bool {
	if _, ok := r.failingSearchFields["download"]; ok {
		return true
	}
	return false
}

func (r *Runner) Download(u string) (io.ReadCloser, http.Header, error) {
	r.createBrowser()

	if required, err := r.isLoginRequired(); required {
		if err := r.login(); err != nil {
			r.logger.WithError(err).Error("Login failed")
			return nil, http.Header{}, err
		}
	} else if err != nil {
		return nil, http.Header{}, err
	}
	//https://arenabg.ch/the-accountant-2016-1080p-brrip-x264-aac-etrg-WuyIajNmV_L0eTjLPZL9GIpkaLom5mB3HcMIxS4eIUw/
	if r.downloadsNeedResolution() {
		//Resolve the url
		downloadItem := r.failingSearchFields["download"]
		r.openPage(u)
		downloadLink, err := r.extractField(r.browser.Dom(), downloadItem)
		if err != nil {
			return nil, nil, err
		}
		u = downloadLink
	}

	fullUrl, err := r.resolveIndexerPath(u)
	if err != nil {
		return nil, http.Header{}, err
	}

	if err := r.browser.Open(fullUrl); err != nil {
		return nil, http.Header{}, err
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

	return pipeR, r.browser.ResponseHeaders(), nil
}
