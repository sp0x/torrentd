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
	facade.IndexScope = NewScope(nil)
	cfg := &config.ViperConfig{}
	cfg.Set("db", tempfile())
	cfg.Set("storage", "boltdb")
	var indexes IndexCollection
	var err error

	go func() {
		indexes, err = facade.IndexScope.LookupAll(cfg, nil)
		g.Expect(err).To(gomega.BeNil())
		close(done)
	}()

	select {
	case <-timeout:
		t.Fatal("test timed out")
	case <-done:
	}
	g.Expect(err).To(gomega.BeNil())
	g.Expect(indexes).ToNot(gomega.BeNil())
}
