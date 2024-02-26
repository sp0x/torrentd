package server

import (
	"github.com/gin-gonic/gin"
	"github.com/sp0x/torrentd/docs"
	"github.com/sp0x/torrentd/server/rss"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) setupRoutes(r *gin.Engine) {
	docs.SwaggerInfo.BasePath = "/"
	// Rss
	r.GET("/all", func(c *gin.Context) { rss.ServerAll(c) })
	r.GET("/movies", func(c *gin.Context) { rss.ServeMovies(c) })
	r.GET("/shows", func(c *gin.Context) { rss.ServeShows(c) })
	r.GET("/music", func(c *gin.Context) { rss.ServeMusic(c) })
	r.GET("/anime", func(c *gin.Context) { rss.ServeAnime(c) })
	r.GET("/search/:name", func(c *gin.Context) {
		rss.SearchAndServe(s.indexerFacade, s.indexerFacade.GetDefaultSearchOptions(), c)
	})
	r.GET("/status", s.Status)
	r.GET("/health", s.HealthCheck)

	// Torznab
	torznab := r.Group("torznab")
	{
		torznab.GET("/", s.torznabHandler)
		torznab.GET("/caps/:indexes", s.torznabIndexCapabilities)
		torznab.GET("/:indexes", s.torznabHandler)
		torznab.GET("/:indexes/api", s.torznabHandler)
	}
	// Aggregated indexers info
	r.GET("t/all/status", s.aggregatesStatus)

	// download routes
	r.HEAD("/download/:token/:filename", func(c *gin.Context) { s.downloadHandler(c) })
	r.GET("/download/:token/:filename", func(c *gin.Context) { s.downloadHandler(c) })
	r.HEAD("/d/:token/:filename", func(c *gin.Context) { s.downloadHandler(c) })
	r.GET("/d/:token/:filename", func(c *gin.Context) { s.downloadHandler(c) })

	url := ginSwagger.URL("/swagger/doc.json") // The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, url))
}
