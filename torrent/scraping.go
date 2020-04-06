package torrent

import (
	"net/url"
	"strconv"
	"strings"
)

func clearSpaces(raw string) string {
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

func stripToNumber(str string) string {
	chars := "0123456789.,"
	var validChars []rune
	for _, c := range []rune(str) {
		if strings.ContainsRune(chars, c) {
			validChars = append(validChars, c)
		}
	}
	str = string(validChars)
	return str
}

func sizeStrToBytes(str string) uint64 {
	str = strings.ToLower(str)
	str = clearSpaces(str)
	multiplier := 1
	if strings.Contains(str, "gb") {
		multiplier = 1028 * 1028 * 1028
	} else if strings.Contains(str, "mb") {
		multiplier = 1028 * 1028
	} else if strings.Contains(str, "kb") {
		multiplier = 1028
	}
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "gb", "", -1)
	str = strings.Replace(str, "mb", "", -1)
	str = strings.Replace(str, "kb", "", -1)
	chars := "1203456789.,"
	var validChars []rune
	for _, c := range []rune(str) {
		if strings.ContainsRune(chars, c) {
			validChars = append(validChars, c)
		}
	}
	str = string(validChars)
	flt, _ := strconv.ParseFloat(str, 32)
	flt = flt * float64(multiplier)
	return uint64(flt)
}
