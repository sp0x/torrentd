package torrent

import (
	"github.com/sp0x/rutracker-rss/db"
	"html"
	"regexp"
	"strings"
)

// Drop any additional info: timestamps, release versions, etc.
// -->
var squareBracesRx, _ = regexp.Compile("^(.+(?:\\s+|\\)))\\[[^\\[\\]]+?\\](.*)$")
var precedingSquareBracesRx, _ = regexp.Compile("^(\\s*)\\[[^\\[\\]]+?\\](.+)$")
var roundBracesRx, _ = regexp.Compile("^(.+(?:\\s+|\\]))\\([^()]+?\\)(.*)$")
var angleBracesRx, _ = regexp.Compile("^(.+)\\s+<<.*?>>(.*)$")
var dateRx, _ = regexp.Compile("^(.+)\\s+(?:\\d{1,2}\\.\\d{1,2}\\.\\d{4}|\\d{4}\\.\\d{2}\\.\\d{2})(.*)$")

// Unable to merge it into date_regex due to some strange behaviour of re
// module.
var date2Rx, _ = regexp.Compile("^(.+)\\s+(?:по|от)\\s+(?:\\d{1,2}\\.\\d{1,2}\\.\\d{4}|\\d{4}\\.\\d{2}\\.\\d{2})(.*)$")
var releaseCounterRx, _ = regexp.Compile("^(.+)\\s+\\d+\\s*(?:в|из)\\s*\\d+(.*)$")
var spacesRx, _ = regexp.Compile("\\s+/.*")
var spaces2Rx, _ = regexp.Compile("\\s+")
var categoriesRx, _ = regexp.Compile("^(national\\s+geographic\\s*:|наука\\s+2\\.0)\\s+")
var arrowsRx, _ = regexp.Compile("^«([^»]{6,})»")
var cyrilicRx, _ = regexp.Compile("^([0-9a-zабвгдеёжзийклмнопрстуфхцчшщьъыэюя., \\-:]{6,}?(?:[:.?!]| - | — |\\|)).*")
var badKeywordsRx, _ = regexp.Compile("(?:\\s|\\()(:?выпуск|выпуски|выпусков|обновлено|передачи за|серия из|сезон|серия|серии|премьера|эфир с|эфир от|эфиры от|satrip)(?:\\s|\\)|$)")

func getTorrentFingerprint(t *db.Torrent) string {
	tagsRx, _ := regexp.Compile("</?[a-z]+>")
	name := strings.Replace(t.Name, "ё", "e", -1)
	name = html.UnescapeString(name)
	name = tagsRx.ReplaceAllString(name, "")

	oldTorrentName := ""
	for name != oldTorrentName {
		oldTorrentName = name
		for _, rx := range []*regexp.Regexp{squareBracesRx, precedingSquareBracesRx, roundBracesRx, angleBracesRx, dateRx,
			date2Rx, releaseCounterRx} {
			name = rx.ReplaceAllString(strings.Trim(name, " .,"), "$1$2")
		}
	}
	name = spacesRx.ReplaceAllString(name, "")
	name = strings.ToLower(name)
	//Shorten it if we can
	name = categoriesRx.ReplaceAllString(name, "")
	name = arrowsRx.ReplaceAllString(name, "$1")
	name = cyrilicRx.ReplaceAllString(name, "$1")
	name = strings.Replace(name, ".", " ", -1)
	//Drop punctuation and other non-alphabet chars
	chars := "abcdefghijklmnopqrstuvwxyzабвгдеёжзийклмнопрстуфхцчшщьъыэюя 123456789+-_.:!,"
	var validatedNameChars []rune
	for _, c := range []rune(name) {
		if strings.ContainsRune(chars, c) {
			validatedNameChars = append(validatedNameChars, c)
		}
	}
	name = string(validatedNameChars)
	name = strings.Replace(name, "г.", "", -1)
	for true {
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
