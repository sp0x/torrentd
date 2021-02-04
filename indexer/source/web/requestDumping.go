package web

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"
)

const dumpFormatHTML = "html"

type dumpData struct {
	ResponseBodyFile string
	RequestBodyFile  string
	State            *jar.State
	Browser          browser.Browsable
}

func (d *dumpData) Write() {
	responseFileWriter, err := os.Create(d.ResponseBodyFile)
	if err != nil {
		logrus.Warnf("could not dump response %s", d.ResponseBodyFile)
		return
	}

	n, err := d.Browser.Download(responseFileWriter)
	if err != nil {
		logrus.Warnf("could not dump response body %s. %v", d.ResponseBodyFile, err)
	} else {
		logrus.Debugf("written response body with size %d bytes", n)
	}
	requestExtension := path.Ext(d.RequestBodyFile)
	if requestExtension != "" && requestExtension != "." {
		requestFileWriter, _ := os.Create(d.RequestBodyFile)
		n := int64(0)
		var err error = nil
		if d.State.Request.Method == "GET" {
			n, err = io.Copy(requestFileWriter, strings.NewReader(d.State.Request.URL.RawQuery))
		} else {
			getBody := d.State.Request.GetBody
			if getBody != nil {
				copyBody, _ := getBody()
				n, err = io.Copy(requestFileWriter, copyBody)
			}
		}
		if err != nil {
			logrus.Warnf("could not dump request body %s", d.RequestBodyFile)
		} else {
			logrus.Debugf("written request body with size %d bytes", n)
		}
	}
}

func (w *Fetcher) dumpFetchData() {
	if !w.options.ShouldDumpData {
		return
	}
	browserState := w.Browser.State()
	request := browserState.Request
	dirPath := path.Join("dumps", request.Host)
	requestURL := request.URL.Path
	dirPath = path.Join(dirPath,
		strings.Replace(fmt.Sprintf("%s_%s", request.Method, requestURL), "/", "_", -1))
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		_ = os.MkdirAll(dirPath, 007)
	}
	timeNow := fmt.Sprintf("%d", time.Now().Unix())
	responseBodyFName := fmt.Sprintf("%s_resp.%s", timeNow, resolveResponseDumpFormat(browserState))
	requestExtension := resolveRequestDumpFormat(browserState)
	requestBodyFName := fmt.Sprintf("%s_req.%s", timeNow, requestExtension)

	responseBodyPath := path.Join(dirPath, responseBodyFName)
	requestBodyPath := path.Join(dirPath, requestBodyFName)
	dump := dumpData{
		ResponseBodyFile: responseBodyPath,
		RequestBodyFile:  requestBodyPath,
		State:            browserState,
		Browser:          w.Browser,
	}
	dump.Write()
}

func resolveRequestDumpFormat(state *jar.State) string {
	if state.Request.Method == "GET" {
		return dumpFormatHTML
	}
	contentType := state.Request.Header.Get("Content-Type")
	if contentType == "" {
		return ""
	}
	return contentTypeToFileExtension(contentType)
}

func resolveResponseDumpFormat(state *jar.State) string {
	contentType := state.Response.Header.Get("Content-Type")
	if contentType == "" {
		return dumpFormatHTML
	}
	return contentTypeToFileExtension(contentType)
}

func contentTypeToFileExtension(fqContentType string) string {
	contentTypeSplit := strings.Split(fqContentType, ";")
	contentType := contentTypeSplit[0]
	switch contentType {
	case "application/json":
		return "json"
	}
	return dumpFormatHTML
}
