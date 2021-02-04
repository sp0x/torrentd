package indexer

import (
	"context"
	"fmt"
	"github.com/sp0x/torrentd/indexer/cache"
	"github.com/sp0x/torrentd/indexer/status"
)

type StatusReporter struct {
	context         context.Context
	indexDefinition *Definition
	errors          cache.LRUCache
}

func (r *StatusReporter) Error(err error) {
	if err == nil {
		return
	}
	status.PublishSchemeError(r.context, generateSchemeErrorStatus(status.LoginError, err, r.indexDefinition))
	errorID := r.errors.Len()
	r.errors.Add(errorID, err)
}

func (r *StatusReporter) GetErrors() []string {
	errs := make([]string, r.errors.Len())
	for i := 0; i < r.errors.Len(); i++ {
		err, ok := r.errors.Get(i)
		if !ok {
			continue
		}
		errs[i] = fmt.Sprintf("%s", err)
	}
	return errs
}
