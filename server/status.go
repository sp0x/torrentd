package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
	"time"
)

var statusCache, _ = cache.NewTTL(10, 3*time.Minute)

type latestResult struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	Site        string `json:"site"`
	Link        string `json:"link"`
}
type statusResponse struct {
	Count  int            `json:"total_count"`
	Latest []latestResult `json:"latest"`
}

func (s *Server) status(c *gin.Context) {
	var statusObj statusResponse
	//If we don't have it in the cache
	if true || !statusCache.Contains("status") {
		store := storage.NewBuilder().
			WithRecord(&search.ExternalResultItem{}).
			Build()
		latest := store.GetLatest(20)
		store.Close()
		var latestResultItems []latestResult
		for _, late := range latest {
			latestResultItems = append(latestResultItems, latestResult{
				Name:        late.Title,
				Description: late.Description,
				Site:        late.Site,
				Link:        late.SourceLink,
			})
		}
		statusObj = statusResponse{
			Count:  len(latest),
			Latest: latestResultItems,
		}
		statusCache.Add("status", statusObj)
	} else {
		cached, _ := statusCache.Get("status")
		statusObj = cached.(statusResponse)
	}
	c.JSON(200, statusObj)
}
