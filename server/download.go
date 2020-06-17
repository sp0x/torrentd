package server

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"github.com/sp0x/torrentd/server/http"
)

func (s *Server) downloadHandler(c http.Context) {
	if c == nil {
		return
	}
	token := c.Param("token")
	filename := c.Param("filename")

	log.WithFields(log.Fields{"filename": filename}).Debugf("Processing download via handler")

	apiKey := s.sharedKey()
	t, err := decodeToken(token, apiKey)
	if err != nil {
		_ = c.Error(err)
		return
	}
	ixr, err := indexer.Lookup(s.config, t.Site)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if t.Link == "" {
		c.String(404, "Indexer link not found")
		return
	}
	rc, err := ixr.Download(t.Link)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if rc == nil {
		_ = c.Error(errors.New("couldn't open stream for download"))
		return
	}
	c.Header("Content-Type", "application/x-bittorrent")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Transfer-Encoding", "binary")
	defer func() {
		_ = rc.Close()
	}()
	c.DataFromReader(200, 0, "application/x-bittorrent", rc, nil)
	//_, _ = io.Copy(c.Writer, rc)
}
