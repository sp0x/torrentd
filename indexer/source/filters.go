package source

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/utils"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sp0x/torrentd/indexer/formatting"
)

var filterLogger log.FieldLogger = log.New()

const (
	filterQueryString = "querystring"
	filterDate        = "dateparse"
	filterTime        = "timeparse"
	filterDateAlt     = "dateparseAlt"
)

type FilterService struct {
}

// Filter out whatever's needed
func (f *FilterService) Filter(fType string, args interface{}, value string) (string, error) {
	switch fType {
	case filterQueryString:
		param, ok := args.(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument", fType)
		}
		return parseQueryString(param, value)

	case filterDate, filterTime, filterDateAlt:
		if args == nil {
			return utils.ParseDate(nil, value)
		}
		if layout, ok := args.(string); ok {
			return utils.ParseDate([]string{layout}, value)
		}
		return "", fmt.Errorf("filter argument type %T was invalid", args)
	case "bool":
		value := formatting.NormalizeSpace(value)
		if value != "" {
			value = "true"
		} else {
			value = "false"
		}
		return value, nil
	case "regexp":
		pattern, ok := args.(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument", fType)
		}
		return filterRegexp(pattern, value)

	case "split":
		sep, ok := (args.([]interface{}))[0].(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument at idx 0", fType)
		}
		pos, ok := (args.([]interface{}))[1].(int)
		if !ok {
			return "", fmt.Errorf("filter %q requires an int argument at idx 1", fType)
		}
		return utils.FilterSplit(sep, pos, value)

	case "replace":
		from, ok := (args.([]interface{}))[0].(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument at idx 0", fType)
		}
		to, ok := (args.([]interface{}))[1].(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument at idx 1", fType)
		}
		return strings.Replace(value, from, to, -1), nil

	case "trim":
		cutset, ok := args.(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument at idx 0", fType)
		}
		return strings.Trim(value, cutset), nil
	case "whitespace":
		return formatting.NormalizeSpace(value), nil
	case "append":
		str, ok := args.(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument at idx 0", fType)
		}
		return value + str, nil

	case "prepend":
		str, ok := args.(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument at idx 0", fType)
		}
		return str + value, nil
	case "urldecode":
		decoded, err := url.QueryUnescape(value)
		if err != nil {
			return "", fmt.Errorf("filter urldecode couldn't decode value `%s`, %s", value, err)
		}
		return decoded, nil
	case "urlarg":
		argName, ok := args.(string)
		if !ok {
			return "", fmt.Errorf("filter %q requires a string argument at idx 0", fType)
		}
		urlx, err := url.Parse(value)
		if err != nil {
			return "", fmt.Errorf("urlarg filter couldn't parse url value: %s", value)
		}
		value = urlx.Query().Get(argName)
		return value, nil
	case "size":
		value = fmt.Sprint(formatting.SizeStrToBytes(value))
		return value, nil
	case "number":
		value = formatting.StripToNumber(formatting.NormalizeSpace(value))
		return value, nil
	case "mapreplace":
		return filterMapReplace(value, args)
	case "re_replace":
		return filterReReplace(value, args)
	case "timeago", "fuzzytime", "reltime":
		return utils.FilterFuzzyTime(value, time.Now(), true)
	}
	return "", errors.New("Unknown filter " + fType)
}

func filterMapReplace(value string, args interface{}) (string, error) {
	replacemenets := args.(map[interface{}]interface{})
	for oldVal, newVal := range replacemenets {
		value = strings.Replace(value, oldVal.(string), newVal.(string), -1)
	}
	return value, nil
}

func filterReReplace(value string, args interface{}) (string, error) {
	pattern := args.([]interface{})[0].(string)
	replacement := args.([]interface{})[1].(string)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	value = re.ReplaceAllString(value, replacement)
	return value, nil
}

func parseQueryString(param string, value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	return u.Query().Get(param), nil
}

func filterRegexp(pattern string, value string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	matches := re.FindStringSubmatch(value)
	if len(matches) == 0 {
		return "", errors.New("no matches found for pattern")
	}
	filterLogger.WithFields(log.Fields{"matches": matches}).Debug("Regex matched")
	// If we have groups, use the groups concatenated by a space
	if len(matches) > 1 {
		matches = matches[1:]
		return strings.Join(matches, " "), nil
	}

	return matches[0], nil
}
