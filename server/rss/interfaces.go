package rss

//HttpContexts allows us to use gin's context
//go:generate mockgen -destination=mocks/mock_httpContext.go -package=mocks . HttpContext
type HttpContext interface {
	//Set a header
	Header(name, value string)
	//Send a string
	String(code int, format string, values ...interface{})
}
