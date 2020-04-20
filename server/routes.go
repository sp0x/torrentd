package server

import "github.com/gin-gonic/gin"

//http://localhost:5000/torznab/rutracker.org?apikey=210fc7bb818639a&t=search&q=bad%20boys
func (s *Server) setupRoutes(r *gin.Engine) {
	//Rss
	r.GET("/all", s.serveAllTorrents)
	r.GET("/movies", s.serveMovies)
	r.GET("/shows", s.serveShows)
	r.GET("/music", s.serveMusic)
	r.GET("/anime", s.serveAnime)
	r.GET("/search/:name", s.searchAndServe)

	//Torznab
	r.GET("torznab/:indexer", s.torznabHandler)
	r.GET("torznab/:indexer/api", s.torznabHandler)

	// download routes
	r.HEAD("/download/:token/:filename", s.downloadHandler)
	r.GET("/download/:token/:filename", s.downloadHandler)
	//r.HEAD("/download/:indexer/:token/:filename", s.downloadHandler)
	//r.GET("/download/:indexer/:token/:filename", s.downloadHandler)
}
