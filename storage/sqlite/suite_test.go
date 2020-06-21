package sqlite

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"testing"
)

func TestSqlite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqlite Suite")
}

// tempfile returns a temporary file path.
func tempfile() string {
	f, err := ioutil.TempFile("", "sqlite-")
	if err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(f.Name()); err != nil {
		panic(err)
	}
	return f.Name()
}
