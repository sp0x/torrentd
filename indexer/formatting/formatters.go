package formatting

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func NormalizeSpace(raw string) string {
	txt := strings.ReplaceAll(raw, "\n", "")
	txt = strings.ReplaceAll(txt, "\t", " ")
	txt = strings.ReplaceAll(txt, "  ", " ")
	return txt
}

func fixMonths(str string) string {
	months := map[string]string{
		"Янв": "Jan",
		"Фев": "Feb",
		"Феб": "Feb",
		"Мар": "Mar",
		"Апр": "Apr",
		"Май": "May",
		"Июн": "Jun",
		"Июл": "Jul",
		"Авг": "Aug",
		"Сен": "Sep",
		"Окт": "Oct",
		"Ноя": "Nov",
		"Дек": "Dec",
	}
	for r, en := range months {
		str = strings.ReplaceAll(str, r, en)
	}
	return str
}

func FormatTime(str string) time.Time {
	// 7-Апр-20 00:06
	str = strings.Trim(str, " \t\n\r")
	str = strings.ReplaceAll(str, "  ", " ")
	str = strings.ReplaceAll(str, "  ", " ")
	str = fixMonths(str)
	parts := strings.Split(str, " ")
	var t time.Time
	var err error
	if len(parts) >= 2 {
		t, err = time.Parse("2-Jan-06 15:04", str)
	} else {
		t, err = time.Parse("2-Jan-06", str)
	}
	if err != nil {
		log.Errorf("Error while parsing time string: %s\t %v\n", str, err)
		return time.Now()
	}
	return t
}

func ExtractAttributeFromQuery(uri string, param string) string {
	furl, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	val := furl.Query().Get(param)
	return val
}

func StripToNumber(str string) string {
	chars := "0123456789.,"
	var validChars []rune
	for _, c := range str {
		if strings.ContainsRune(chars, c) {
			validChars = append(validChars, c)
		}
	}
	str = string(validChars)
	return str
}

func SizeStrToBytes(str string) uint64 {
	str = strings.ToLower(str)
	str = NormalizeSpace(str)
	multiplier := 1
	switch {
	case strings.Contains(str, "gb"):
		multiplier = 1028 * 1028 * 1028
	case strings.Contains(str, "mb"):
		multiplier = 1028 * 1028
	case strings.Contains(str, "kb"):
		multiplier = 1028
	}
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, "gb", "")
	str = strings.ReplaceAll(str, "mb", "")
	str = strings.ReplaceAll(str, "kb", "")
	chars := "1203456789.,"
	var validChars []rune
	for _, c := range str {
		if strings.ContainsRune(chars, c) {
			validChars = append(validChars, c)
		}
	}
	str = string(validChars)
	flt, _ := strconv.ParseFloat(str, 32)
	flt *= float64(multiplier)
	return uint64(flt)
}
