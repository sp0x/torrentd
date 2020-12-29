package indexer

import (
	"testing"
	"time"
)

func TestCachedScope_CreateAggregate_ShouldNotHang(t *testing.T) {
	g := gomega.NewWithT(t)
	timeout := time.After(30 * time.Second)
	done := make(chan struct{})
	facade := Facade{}
	facade.Scope = NewScope()
	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	_ = cfg.Set("storage", "boltdb")
	var aggregate Indexer
	var err error

	go func() {
		aggregate, err = facade.Scope.CreateAggregate(cfg, nil)
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
