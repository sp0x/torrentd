package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/storage"
	"time"
)

var statusCache, _ = cache.NewTTL(10, 3*time.Minute)

func (s *Server) status(c *gin.Context) {
	var statusObj interface{}
	//If we don't have it in the cache
	if !statusCache.Contains("status") {
		strg := storage.DefaultStorage()
		latest := strg.GetNewest(10)
		var latestNames = []string{}
		for _, late := range latest {
			latestNames = append(latestNames, late.Title)
		}
		statusObj = struct {
			Torrents int64    `json:"total_count"`
			Latest   []string `json:"latest"`
		}{
			Torrents: strg.GetTorrentCount(),
			Latest:   latestNames,
		}
		statusCache.Add("status", statusObj)
	} else {
		statusObj, _ = statusCache.Get("status")
	}
	c.JSON(200, statusObj)
}