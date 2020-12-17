package server

import "github.com/gin-gonic/gin"

type healthCheckResponse struct {
	Ok    bool  `json:"ok"`
	Error error `json:"error,omitempty"`
}

func (s *Server) HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		indexes := s.indexerFacade.Scope.Indexes()
		output := make(map[string]healthCheckResponse)
		for _, index := range indexes {
			err := index.HealthCheck()
			output[index.Site()] = healthCheckResponse{
				Ok:    err == nil,
				Error: err,
			}
		}
		c.JSON(200, output)
	}
}
