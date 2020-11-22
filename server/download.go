package server

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/server/http"
)

func (s *Server) downloadHandler(c http.Context) {
	if c == nil {
		return
	}
	token := c.Param("token")
	filename := c.Param("filename")
	if token == "" {
		return
	}
	log.WithFields(log.Fields{"filename": filename}).Debugf("Processing download via handler")

	apiKey := s.sharedKey()
	t, err := decodeToken(token, apiKey)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if t.Link == "" {
		c.String(404, "Indexer link not found")
		return
	}
	index, err := s.indexerFacade.Scope.Lookup(s.config, t.IndexName)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if index == nil {
		_ = c.Error(errors.New("indexer not found"))
		return
	}
	downloadProxy, err := index.Download(t.Link)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if downloadProxy == nil || nil == downloadProxy.Reader {
		_ = c.Error(errors.New("couldn't open stream for download"))
		return
	}
	defer func() {
		_ = downloadProxy.Reader.Close()
	}()
	log.WithFields(log.Fields{"link": t.Link}).
		Infof("Waiting for download")
	c.Header("Content-Type", "application/x-bittorrent")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.DataFromReader(200, 0, "application/x-bittorrent", downloadProxy.Reader, nil)
	//select {
	//case length := <-downloadProxy.ContentLengthChan:
	//	c.Header("Content-Type", "application/x-bittorrent")
	//	c.Header("Content-Disposition", "attachment; filename="+filename)
	//	c.Header("Content-Transfer-Encoding", "binary")
	//	c.DataFromReader(200, length, "application/x-bittorrent", downloadProxy.Reader, nil)
	//case <-time.After(20 * time.Second):
	//	c.String(408, "Timed out waiting for download")
	//}

}
