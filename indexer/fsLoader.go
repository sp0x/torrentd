package indexer

import (
	"github.com/sp0x/rutracker-rss/config"
	"os"
	"path"
	"strings"
)

type fsLoader struct {
	dirs []string
}

func newFsLoader() DefinitionLoader {
	return &fsLoader{config.GetDefinitionDirs()}
}

func (fs *fsLoader) walkDirectories() (map[string]string, error) {
	defs := map[string]string{}

	for _, dirpath := range fs.dirs {
		dir, err := os.Open(dirpath)
		if os.IsNotExist(err) {
			continue
		}
		files, err := dir.Readdirnames(-1)
		if err != nil {
			continue
		}
		for _, basename := range files {
			if strings.HasSuffix(basename, ".yml") || strings.HasSuffix(basename, ".yaml") {
				index := strings.TrimSuffix(basename, ".yml")
				index = strings.TrimSuffix(index, ".yaml")
				indexFile := path.Join(dir.Name(), basename)
				defs[index] = path.Join(indexFile)
			}
		}
	}

	return defs, nil
}

func (fs *fsLoader) List() ([]string, error) {
	defs, err := fs.walkDirectories()
	if err != nil {
		return nil, err
	}
	var results []string
	for k := range defs {
		results = append(results, k)
	}
	return results, nil
}

func (fs *fsLoader) String() string {
	buff := ""
	defs := fs.dirs
	for _, def := range defs {
		buff += def + "\n"
	}
	return buff
}

//Load - Load a definition of an Indexer from it's name
func (fs *fsLoader) Load(key string) (*IndexerDefinition, error) {
	defs, err := fs.walkDirectories()
	if err != nil {
		return nil, err
	}

	fileName, ok := defs[key]
	if !ok {
		return nil, ErrUnknownIndexer
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	def, err := ParseDefinitionFile(f)
	if err != nil {
		return def, err
	}

	def.stats.Source = "file:" + fileName
	return def, err
}
