package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/indexer"
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

	indexer, err := indexer.Lookup(s.config, t.Site)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusBadGateway)
		return
	}
	if t.Link == "" {
		c.Writer.WriteString("Indexer link not found.\n")
		http.NotFound(c.Writer, c.Request)
		return
	}
	rc, err := indexer.Download(t.Link)
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
