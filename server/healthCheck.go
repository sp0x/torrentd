package server

import "github.com/gin-gonic/gin"

type indexHealthCheckResponse struct {
	Ok    bool  `json:"ok"`
	Error error `json:"error,omitempty"`
}

func (s *Server) HealthCheck(c *gin.Context) {
	indexes := s.indexerFacade.IndexScope.Indexes()
	output := make(map[string]indexHealthCheckResponse)
	for _, indexGroup := range indexes {
		err := indexGroup.HealthCheck()
		firstIndex := indexGroup[0]

		output[firstIndex.Site()] = indexHealthCheckResponse{
			Ok:    err == nil,
			Error: err,
		}
	}
	c.JSON(200, output)
}
