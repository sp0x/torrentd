package indexer

import (
	"github.com/onsi/gomega"
	"github.com/sp0x/torrentd/config"
	"testing"
)

func newTestingIndex() *Runner {
	cfg := &config.ViperConfig{}
	_ = cfg.Set("db", tempfile())
	_ = cfg.Set("storage", "boltdb")
	runner := NewRunner(index, RunnerOpts{
		Config:     cfg,
		CachePages: false,
		Transport:  nil,
	})
	return runner
}

func TestNewSessionMultiplexer(t *testing.T) {
	g := gomega.NewWithT(t)
	sampleIndex := newTestingIndex()

	multiplexer, err := NewSessionMultiplexer(sampleIndex, 10)

	g.Expect(err).To(gomega.BeNil())
	g.Expect(len(multiplexer.sessions)).To(gomega.Equal(10))
}
