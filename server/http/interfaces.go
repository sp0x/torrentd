package http

import (
	"github.com/gin-gonic/gin"
	"io"
)

//HttpContexts allows us to use gin's context
//go:generate mockgen -destination=mocks/mock_httpContext.go -package=mocks . HttpContext
type Context interface {
	//Set a header
	Header(name, value string)
	//Send a string
	String(code int, format string, values ...interface{})
	Param(s string) string
	Error(err error) *gin.Error
	DataFromReader(code int, contentLength int64, contentType string, reader io.Reader, extraHeaders map[string]string)
}
