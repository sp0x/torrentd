package indexer

import (
	"github.com/sirupsen/logrus"
	"github.com/sp0x/rutracker-rss/config"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

//If caching is enabled, we cache the page's contents in our pagecache
func (r *Runner) cachePage() error {
	if !r.opts.CachePages {
		return nil
	}
	dir := config.GetCachePath(r.definition.Site)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmpfile, err := ioutil.TempFile(dir, "pagecache")
	if err != nil {
		r.logger.Warn(err)
		return err
	}

	body := strings.NewReader(r.browser.Body())
	_, _ = io.Copy(tmpfile, body)
	if err = tmpfile.Close(); err != nil {
		return err
	}

	newFile := tmpfile.Name() + ".html"
	_ = os.Rename(tmpfile.Name(), newFile)

	r.logger.
		WithFields(logrus.Fields{"file": "file://" + newFile}).
		Debugf("Wrote page output to cache")

	return nil
}
