package indexer

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"

	"github.com/sp0x/torrentd/config"
)

type FileIndexLoader struct {
	Directories []string
}

// NewFsLoader creates a new file index loader which looks for definitions in ~/.#{appName}/indexes and ./indexes
func NewFsLoader(appName string) *FileIndexLoader {
	localDirectory := ""
	entityName := "indexes"
	if cwd, err := os.Getwd(); err == nil {
		localDirectory = filepath.Join(cwd, entityName)
	}
	home, _ := homedir.Dir()
	homeDefsDir := path.Join(home, fmt.Sprintf(".%s", appName), entityName)
	x := &FileIndexLoader{
		Directories: []string{
			localDirectory,
			homeDefsDir,
		},
	}
	return x
}

func defaultFsLoader() DefinitionLoader {
	return &FileIndexLoader{config.GetDefinitionDirs()}
}

func (fs *FileIndexLoader) walkDirectories() (map[string]string, error) {
	defs := map[string]string{}

	for _, dirpath := range fs.Directories {
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

func (fs *FileIndexLoader) List(selector *IndexerSelector) ([]string, error) {
	defs, err := fs.walkDirectories()
	if err != nil {
		return nil, err
	}
	var results []string
	for name := range defs {
		if selector != nil && !selector.Matches(name) {
			continue
		}
		results = append(results, name)
	}
	return results, nil
}

func (fs *FileIndexLoader) ListWithNames(names []string) ([]string, error) {
	defs, err := fs.walkDirectories()
	if err != nil {
		return nil, err
	}
	var results []string
	for k := range defs {
		if !contains(names, k) {
			continue
		}
		results = append(results, k)
	}
	return results, nil
}

func (fs *FileIndexLoader) String() string {
	buff := ""
	defs := fs.Directories
	for _, def := range defs {
		buff += def + "\n"
	}
	return "dirs{" + buff + "}"
}

// Load - Load a definition of an Indexer from it's name
func (fs *FileIndexLoader) Load(key string) (*IndexerDefinition, error) {
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
