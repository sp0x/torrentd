package formatting

import (
	"html"
	"regexp"
	"strings"
)

// Drop any additional info: timestamps, release versions, etc.
// -->
var (
	squareBracesRx          = regexp.MustCompile(`^(.+(?:\s+|\)))\[[^\[\]]+?\](.*)$`)
	precedingSquareBracesRx = regexp.MustCompile(`^(\s*)\[[^\[\]]+?\](.+)$`)
	roundBracesRx           = regexp.MustCompile(`^(.+(?:\s+|\]))\([^()]+?\)(.*)$`)
	angleBracesRx           = regexp.MustCompile(`^(.+)\s+<<.*?>>(.*)$`)
	dateRx                  = regexp.MustCompile(`^(.+)\s+(?:\d{1,2}\.\d{1,2}\.\d{4}|\d{4}\.\d{2}\.\d{2})(.*)$`)
)

// Unable to merge it into date_regex due to some strange behaviour of re
// module.
var (
	date2Rx          = regexp.MustCompile(`^(.+)\s+(?:по|от)\s+(?:\d{1,2}\.\d{1,2}\.\d{4}|\d{4}\.\d{2}\.\d{2})(.*)$`)
	releaseCounterRx = regexp.MustCompile(`^(.+)\s+\d+\s*(?:в|из)\s*\d+(.*)$`)
	spacesRx         = regexp.MustCompile(`\s+/.*`)
	spaces2Rx        = regexp.MustCompile(`\s+`)
	categoriesRx     = regexp.MustCompile(`^(national\s+geographic\s*:|наука\s+2\.0)\s+`)
	arrowsRx         = regexp.MustCompile("^«([^»]{6,})»")
	cyrilicRx        = regexp.MustCompile(`^([0-9a-zабвгдеёжзийклмнопрстуфхцчшщьъыэюя., \-:]{6,}?(?:[:.?!]| - | — |\|)).*`)
	badKeywordsRx    = regexp.MustCompile(`(?:\s|\()(:?выпуск|выпуски|выпусков|обновлено|передачи за|серия из|сезон|серия|серии|премьера|эфир с|эфир от|эфиры от|satrip)(?:\s|\)|$)`)
)

func GetResultFingerprint(title string) string {
	tagsRx := regexp.MustCompile("</?[a-z]+>")
	name := strings.ReplaceAll(title, "ё", "e")
	name = html.UnescapeString(name)
	name = tagsRx.ReplaceAllString(name, "")

	oldTorrentName := ""
	for name != oldTorrentName {
		oldTorrentName = name
		for _, rx := range []*regexp.Regexp{
			squareBracesRx, precedingSquareBracesRx, roundBracesRx, angleBracesRx, dateRx,
			date2Rx, releaseCounterRx,
		} {
			name = rx.ReplaceAllString(strings.Trim(name, " .,"), "$1$2")
		}
	}
	name = spacesRx.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	// Shorten it if we can
	name = categoriesRx.ReplaceAllString(name, "")
	name = arrowsRx.ReplaceAllString(name, "$1")
	name = cyrilicRx.ReplaceAllString(name, "$1")
	name = strings.ReplaceAll(name, ".", " ")
	// Drop punctuation and other non-alphabet chars
	chars := "abcdefghijklmnopqrstuvwxyzабвгдеёжзийклмнопрстуфхцчшщьъыэюя 123456789+-_.:!,"
	var validatedNameChars []rune
	for _, c := range name {
		if strings.ContainsRune(chars, c) {
			validatedNameChars = append(validatedNameChars, c)
		}
	}
	name = string(validatedNameChars)
	name = strings.ReplaceAll(name, "г.", "")
	for {
		newName := badKeywordsRx.ReplaceAllString(name, "")
		if newName == name {
			break
		}
		name = newName
	}
	for _, month := range []string{
		"январь", "января",
		"февраль", "февраля",
		"март", "марта",
		"апрель", "апреля",
		"май", "мая",
		"июнь", "июня",
		"июль", "июля",
		"август", "августа",
		"сентябрь", "сентября",
		"октябрь", "октября",
		"ноябрь", "ноября",
		"декабрь", "декабря",
	} {
		monthRx, _ := regexp.Compile("\b" + month + "\b")
		name = monthRx.ReplaceAllString(name, "")
	}

	name = spaces2Rx.ReplaceAllString(name, " ")
	return name
}
