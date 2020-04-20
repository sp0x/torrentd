package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

func (s *Server) downloadHandler(c *gin.Context) {
	_ = c.Params
	token := c.Param("token")
	filename := c.Param("filename")

	log.WithFields(log.Fields{"filename": filename}).Debugf("Processing download via handler")

	k, err := s.sharedKey()
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := decodeToken(token, k)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadGateway)
		return
	}

	indexer, err := s.lookupIndexer(t.Site)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadGateway)
		return
	}

	rc, _, err := indexer.Download(t.Link)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadGateway)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/x-bittorrent")
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Writer.Header().Set("Content-Transfer-Encoding", "binary")

	defer rc.Close()
	io.Copy(c.Writer, rc)
}
