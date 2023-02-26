package server

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/status/models"
)

var cacheTTL = 3 * time.Minute
var statusCache, _ = cache.NewTTL(10, cacheTTL)

type statusResponse struct {
	Latest  []models.LatestResult `json:"latest"`
	Indexes []models.IndexStatus  `json:"indexes"`
}

// Status godoc
// @Summary      Status of the server
// @Description  get status of the server
// @Tags         status
// @Accept       */*
// @Produce      json
// @Success      200  {object}  statusResponse
// @Router       /status [get]
func (s *Server) Status(c *gin.Context) {
	var statusObj statusResponse

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
