package main

import (
	"net/url"
	"strings"
)

func cleanupHtmlText(raw string) string {
	txt := strings.Replace(raw, "\n", "", -1)
	txt = strings.Replace(txt, "\t", "  ", -1)
	txt = strings.Replace(txt, "  ", " ", -1)
	return txt
}

func extractAttr(uri string, param string) string {
	furl, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	val := furl.Query().Get(param)
	return val
}

func sizeStrToBytes(str string) uint64 {
	str = strings.ToLower(str)
	str = strings.Replace(str, "gb", "")
	return 0
}
