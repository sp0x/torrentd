package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/status/models"
	"time"
)

var statusCache, _ = cache.NewTTL(10, 3*time.Minute)

type statusResponse struct {
	Latest  []models.LatestResult `json:"latest"`
	Indexes []models.IndexStatus  `json:"indexes"`
}

func (s *Server) Status(c *gin.Context) {
	var statusObj statusResponse
	//If we don't have it in the cache
	if !statusCache.Contains("status") {
		latestResultItems := s.status.GetLatestItems()
		statusObj = statusResponse{
			Latest: latestResultItems,
		}
		statusCache.Add("status", statusObj)
	} else {
		cached, _ := statusCache.Get("status")
		statusObj = cached.(statusResponse)
	}
	statusObj.Indexes = s.status.GetIndexesStatus(s.indexerFacade)
	c.JSON(200, statusObj)
}
