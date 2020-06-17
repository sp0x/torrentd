package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sp0x/torrentd/server/rss"
)

//http://localhost:5000/torznab/rutracker.org?apikey=210fc7bb818639a&t=search&q=bad%20boys
func (s *Server) setupRoutes(r *gin.Engine) {
	//Rss
	r.GET("/all", func(c *gin.Context) { rss.ServerAll(c) })
	r.GET("/movies", func(c *gin.Context) { rss.ServeMovies(c) })
	r.GET("/shows", func(c *gin.Context) { rss.ServeShows(c) })
	r.GET("/music", func(c *gin.Context) { rss.ServeMusic(c) })
	r.GET("/anime", func(c *gin.Context) { rss.ServeAnime(c) })
	r.GET("/search/:name", func(c *gin.Context) {
		rss.SearchAndServe(s.tracker, s.tracker.GetDefaultOptions(), c)
	})
	r.GET("/status", s.status)

	//Torznab
	r.GET("torznab/:indexer", s.torznabHandler)
	r.GET("torznab/:indexer/api", s.torznabHandler)
	//Aggregated indexers info
	r.GET("t/all/status", s.aggregatesStatus)

	// download routes
	r.HEAD("/download/:token/:filename", func(c *gin.Context) { s.downloadHandler(c) })
	r.GET("/download/:token/:filename", func(c *gin.Context) { s.downloadHandler(c) })
	r.HEAD("/d/:token/:filename", func(c *gin.Context) { s.downloadHandler(c) })
	r.GET("/d/:token/:filename", func(c *gin.Context) { s.downloadHandler(c) })
}
