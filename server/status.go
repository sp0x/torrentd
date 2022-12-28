package server

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/status/models"
)

var statusCache, _ = cache.NewTTL(10, 3*time.Minute)

type statusResponse struct {
	Latest  []models.LatestResult `json:"latest"`
	Indexes []models.IndexStatus  `json:"indexes"`
}

// Status godoc
// @Summary      Show an account
// @Description  get status of the server
// @Tags         status
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.Account
// @Failure      400  {object}  httputil.HTTPError
// @Failure      404  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
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
