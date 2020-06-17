package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer"
	"io"
	"net/http"
)

func (s *Server) downloadHandler(c *gin.Context) {
	_ = c.Params
	token := c.Param("token")
	filename := c.Param("filename")

	log.WithFields(log.Fields{"filename": filename}).Debugf("Processing download via handler")

	apiKey := s.sharedKey()
	t, err := decodeToken(token, apiKey)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadGateway)
		return
	}

	ixr, err := indexer.Lookup(s.config, t.Site)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadGateway)
		return
	}
	if t.Link == "" {
		_, _ = c.Writer.WriteString("Indexer link not found.\n")
		http.NotFound(c.Writer, c.Request)
		return
	}
	rc, err := ixr.Download(t.Link)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadGateway)
		return
	}
	if rc == nil {
		http.Error(c.Writer, "Couldn't open stream for download", http.StatusBadGateway)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/x-bittorrent")
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Writer.Header().Set("Content-Transfer-Encoding", "binary")

	defer func() {
		_ = rc.Close()
	}()
	_, _ = io.Copy(c.Writer, rc)
}
