package indexer

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

var escFilenameRegex = regexp.MustCompile(`^/definitions/(.+?)\.yml$`)

type escLoader struct {
	http.FileSystem
}

func (el escLoader) List() ([]string, error) {
	results := []string{}

	for _, filename := range []string{} {
		if matches := escFilenameRegex.FindStringSubmatch(filename); matches != nil {
			results = append(results, matches[1])
		}
	}

	return results, nil
}

func (el escLoader) Load(key string) (*IndexerDefinition, error) {
	fname := fmt.Sprintf("/definitions/%s.yml", key)
	f, err := el.Open(fname)
	if os.IsNotExist(err) {
		return nil, ErrUnknownIndexer
	} else if err != nil {
		return nil, err
	}

	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	def, err := ParseDefinition(data)
	if err != nil {
		return def, err
	}

	fi, err := f.Stat()
	if err != nil {
		return def, err
	}

	def.stats.ModTime = fi.ModTime()
	def.stats.Source = "builtin:" + fname
	return def, nil
}
