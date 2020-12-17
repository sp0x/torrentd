package web

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/surf/jar"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

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
		getBody := d.State.Request.GetBody
		if getBody != nil {
			copyBody, _ := getBody()
			n, err := io.Copy(requestFileWriter, copyBody)
			if err != nil {
				logrus.Warnf("could not dump request body %s", d.RequestBodyFile)
			} else {
				logrus.Debugf("written request body with size %d bytes", n)
			}
		}
	}
}

func (w *ContentFetcher) dumpFetchData() {
	if !w.options.DumpData {
		return
	}
	browserState := w.Browser.State()
	request := browserState.Request
	dirPath := path.Join("dumps", request.Host)
	requestUrl := request.URL.Path + "__" + request.URL.RawQuery
	dirPath = path.Join(dirPath,
		strings.Replace(fmt.Sprintf("%s_%s", request.Method, requestUrl), "/", "_", -1))
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		os.MkdirAll(dirPath, 007)
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
		return ""
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
		return "html"
	}
	return contentTypeToFileExtension(contentType)
}

func contentTypeToFileExtension(contentType string) string {
	switch contentType {
	case "application/json":
		return "json"
	}
	return "html"
}
