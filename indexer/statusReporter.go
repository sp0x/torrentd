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

type ReportableError struct {
	error
	Type string
}

func (r *ReportableError) Error() string {
	return r.error.Error()
}

func NewError(errType string, err error) *ReportableError {
	return &ReportableError{
		Type:  errType,
		error: err,
	}
}

func (r *StatusReporter) Error(err *ReportableError) {
	if err == nil {
		return
	}
	status.PublishSchemeError(r.context, generateSchemeErrorStatus(err.Type, err.error, r.indexDefinition))

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
