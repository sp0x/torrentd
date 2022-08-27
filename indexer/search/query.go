package search

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/sp0x/torrentd/indexer/categories"
)

type Query struct {
	Type                                         string
	QueryString, Series, Ep, Season, Movie, Year string
	Limit, Offset                                int
	Extended                                     bool
	Categories                                   []int
	APIKey                                       string

	// identifier types
	TVDBID       string
	TVRageID     string
	IMDBID       string
	TVMazeID     string
	TraktID      string
	Fields       map[string]interface{}
	StopOnStale          bool
	NumberOfPagesToFetch uint
	//StartingPage uint
	Page         uint
}

func NewQuery() *Query {
	q := &Query{}
	q.Fields = make(map[string]interface{})
	return q
}

func NewQueryFromQueryString(query string) (*Query, error) {
	q := NewQuery()
	q.Type = "search"
	q.Fields = make(map[string]interface{})
	if queryUsesPatterns(query) {
		err := parsePatternQuery(q, query)
		if err != nil {
			return nil, err
		}
	} else {
		q.QueryString = query
	}
	return q, nil
}

func parsePatternQuery(q *Query, rawQuery string) error {
	patterns := strings.Split(rawQuery, ";")
	for _, part := range patterns {
		partSplit := strings.SplitN(part, ":", 2)
		field := partSplit[0][1:]
		fieldValue := partSplit[1]
		if function := parseQueryPatternValue(fieldValue); function != nil {
			err := populateFieldWithQueryFunction(q, field, function)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type queryFunction struct {
	name   string
	params []string
}

func trimSpaces(params []string) []string {
	for i, val := range params {
		params[i] = strings.Trim(val, " \t\n")
	}
	return params
}

func parseQueryPatternValue(patternValue string) *queryFunction {
	if strings.Contains(patternValue, "(") && strings.Contains(patternValue, ")") {
		split := strings.SplitN(patternValue, "(", 2)
		funcName := split[0]
		if funcName == "range" {
			paramsStr := split[1][0 : len(split[1])-1]
			params := strings.Split(paramsStr, ",")
			params = trimSpaces(params)
			return &queryFunction{
				name:   funcName,
				params: params,
			}
		}
	}
	return nil
}

func populateFieldWithQueryFunction(q *Query, field string, function *queryFunction) error {
	if function.name == "range" {
		q.Fields[field] = RangeField(function.params)
	} else {
		return fmt.Errorf("function `%s` is not supported", function.name)
	}

	return nil
}

func queryUsesPatterns(query string) bool {
	patterns := strings.Split(query, ";")
	for _, part := range patterns {
		if strings.HasPrefix(part, "$") {
			return true
		}
	}
	return false
}

// ParseQuery takes the query string parameters for a torznab query and parses them
func ParseQuery(v url.Values) (*Query, error) {
	query := &Query{}

	for k, vals := range v {
		switch k {
		case "t":
			if len(vals) > 1 {
				return query, errors.New("multiple t parameters not allowed")
			}
			query.Type = vals[0]
		case "p":
			cnt, _ := strconv.Atoi(vals[0])
			query.NumberOfPagesToFetch = uint(cnt)
		case "q":
			query.QueryString = strings.Join(vals, " ")

		case "series":
			query.Series = strings.Join(vals, " ")

		case "movie":
			query.Movie = strings.Join(vals, " ")

		case "year":
			if len(vals) > 1 {
				return query, errors.New("multiple year parameters not allowed")
			}
			query.Year = vals[0]

		case "ep":
			if len(vals) > 1 {
				return query, errors.New("multiple ep parameters not allowed")
			}
			query.Ep = vals[0]

		case "season":
			if len(vals) > 1 {
				return query, errors.New("multiple season parameters not allowed")
			}
			query.Season = vals[0]

		case "apikey":
			if len(vals) > 1 {
				return query, errors.New("multiple apikey parameters not allowed")
			}
			query.APIKey = vals[0]

		case "limit":
			if len(vals) > 1 {
				return query, errors.New("multiple limit parameters not allowed")
			}
			limit, err := strconv.Atoi(vals[0])
			if err != nil {
				return query, err
			}
			query.Limit = limit

		case "offset":
			if len(vals) > 1 {
				return query, errors.New("multiple offset parameters not allowed")
			}
			offset, err := strconv.Atoi(vals[0])
			if err != nil {
				return query, err
			}
			query.Offset = offset

		case "extended":
			if len(vals) > 1 {
				return query, errors.New("multiple extended parameters not allowed")
			}
			extended, err := strconv.ParseBool(vals[0])
			if err != nil {
				return query, err
			}
			query.Extended = extended

		case "cat":
			query.Categories = []int{}
			for _, val := range vals {
				ints, err := splitInts(val, ",")
				if err != nil {
					return nil, fmt.Errorf("unable to parse cats %q", vals[0])
				}
				query.Categories = append(query.Categories, ints...)
			}

		case "format":

		case "tvdbid":
			if len(vals) > 1 {
				return query, errors.New("multiple tvdbid parameters not allowed")
			}
			query.TVDBID = vals[0]

		case "rid":
			if len(vals) > 1 {
				return query, errors.New("multiple rid parameters not allowed")
			}
			query.TVRageID = vals[0]

		case "tvmazeid":
			if len(vals) > 1 {
				return query, errors.New("multiple tvmazeid parameters not allowed")
			}
			query.TVMazeID = vals[0]

		case "imdbid":
			if len(vals) > 1 {
				return query, errors.New("multiple imdbid parameters not allowed")
			}
			query.IMDBID = vals[0]

		default:
			log.Warningf("Unknown torznab request key %q\n", k)
		}
	}

	return query, nil
}

func splitInts(s, delim string) (i []int, err error) {
	for _, v := range strings.Split(s, delim) {
		vInt, err := strconv.Atoi(v)
		if err != nil {
			return i, err
		}
		i = append(i, vInt)
	}
	return i, err
}

// Episode returns either the season + episode in the format S00E00 or just the season as S00 if
// no episode has been specified.
func (query *Query) Episode() (s string) {
	if query.Season != "" {
		s += fmt.Sprintf("S%02s", query.Season)
	}
	if query.Ep != "" {
		s += fmt.Sprintf("E%02s", query.Ep)
	}
	return s
}

// AddCategory adds a category to the query
func (query *Query) AddCategory(cat categories.Category) {
	if query.Categories == nil {
		query.Categories = []int{}
	}
	query.Categories = append(query.Categories, cat.ID)
}

func (query *Query) ToVerboseString() string {
	output := ""
	keywords := query.Keywords()
	var data []string
	if keywords != "" {
		data = append(data, "keywords: "+keywords)
	}
	fieldsStr := fieldsToString(query)
	if fieldsStr != "" {
		data = append(data, "fields: "+fieldsStr)
	}
	output = strings.Join(data, "; ")
	return output
}

func fieldsToString(q *Query) string {
	output := []string{}
	for fname, fval := range q.Fields {
		val := fmt.Sprintf("{%s: %v}", fname, fval)
		output = append(output, val)
	}
	return strings.Join(output, ", ")
}

// Keywords returns the query formatted as search keywords
func (query *Query) Keywords() string {
	tokens := []string{}

	if query.QueryString != "" {
		tokens = append(tokens, query.QueryString)
	}

	if query.Series != "" {
		tokens = append(tokens, query.Series)
	}

	if query.Movie != "" {
		tokens = append(tokens, query.Movie)
	}

	if query.Year != "" {
		tokens = append(tokens, query.Year)
	}

	if query.Season != "" || query.Ep != "" {
		tokens = append(tokens, query.Episode())
	}

	return strings.Join(tokens, " ")
}

// Encode returns the query as a url query string
func (query *Query) Encode() string {
	v := url.Values{}

	if query.Type != "" {
		v.Set("t", query.Type)
	} else {
		v.Set("t", "search")
	}

	if query.QueryString != "" {
		v.Set("q", query.QueryString)
	}

	if query.Ep != "" {
		v.Set("ep", query.Ep)
	}

	if query.Season != "" {
		v.Set("season", query.Season)
	}

	if query.Movie != "" {
		v.Set("movie", query.Movie)
	}

	if query.Year != "" {
		v.Set("year", query.Year)
	}

	if query.Series != "" {
		v.Set("series", query.Series)
	}

	if query.Offset != 0 {
		v.Set("offset", strconv.Itoa(query.Offset))
	}

	if query.Limit != 0 {
		v.Set("limit", strconv.Itoa(query.Limit))
	}

	if query.Extended {
		v.Set("extended", "1")
	}

	if query.APIKey != "" {
		v.Set("apikey", query.APIKey)
	}

	if len(query.Categories) > 0 {
		cats := []string{}

		for _, cat := range query.Categories {
			cats = append(cats, strconv.Itoa(cat))
		}

		v.Set("cat", strings.Join(cats, ","))
	}

	if query.TVDBID != "" {
		v.Set("tvdbid", query.TVDBID)
	}

	if query.TVRageID != "" {
		v.Set("rid", query.TVRageID)
	}

	if query.TVMazeID != "" {
		v.Set("tvmazeid", query.TVMazeID)
	}

	if query.TraktID != "" {
		v.Set("traktid", query.TraktID)
	}

	if query.IMDBID != "" {
		v.Set("imdbid", query.IMDBID)
	}

	return v.Encode()
}

func (query *Query) String() string {
	return query.Encode()
}

func (query *Query) UniqueKey() interface{} {
	encoded := query.Encode()
	return encoded
}

func (query *Query) HasEnoughResults(numberOfResults int) bool {
	return query.Limit > 0 && numberOfResults >= query.Limit
}
