package indexer

import (
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/sp0x/torrentd/config"
)

func TestIndexCollection_CreateAggregate_ShouldNotHang(t *testing.T) {
	g := gomega.NewWithT(t)
	timeout := time.After(5 * time.Second)
	done := make(chan struct{})
	facade := Facade{}
	facade.LoadedIndexes = NewScope()
	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	_ = cfg.Set("storage", "boltdb")
	var aggregate Indexer
	var err error

	go func() {
		aggregate, err = facade.LoadedIndexes.CreateAggregate(cfg, nil)
		g.Expect(err).To(gomega.BeNil())
		close(done)
	}()

	select {
	case <-timeout:
		t.Fatal("test timed out")
	case <-done:
	}
	g.Expect(err).To(gomega.BeNil())
	g.Expect(aggregate).ToNot(gomega.BeNil())
}
