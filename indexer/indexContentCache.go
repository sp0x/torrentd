package indexer

import (
	"github.com/sirupsen/logrus"
	"github.com/sp0x/surf/browser"
	"github.com/sp0x/torrentd/config"
	"io/ioutil"
	"os"
)

//If caching is enabled, we cache the page's contents in our pagecache
//the current browser page is cached
func (r *Runner) CachePage(browsable browser.Browsable) error {
	if !r.opts.CachePages {
		return nil
	}
	//Store the cache in directories based on the index definition.
	dir := config.GetCachePath(r.definition.Site)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	tmpfile, err := ioutil.TempFile(dir, "pagecache")
	if err != nil {
		r.logger.Warn(err)
		return err
	}
	_, _ = browsable.Download(tmpfile)
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
