package indexer

import (
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/status/models"
	"github.com/sp0x/torrentd/storage"
)

type StandardReportGenerator struct{}

//go:generate mockgen -source reportGenerator.go -destination=mocks/reportGenerator.go -package=mocks
type ReportGenerator interface {
	GetLatestItems() []models.LatestResult
	GetIndexesStatus(s *Facade) []models.IndexStatus
}

func (st *StandardReportGenerator) GetLatestItems() []models.LatestResult {
	store := storage.NewBuilder().
		WithRecord(&search.ExternalResultItem{}).
		Build()
	latest := store.GetLatest(20)
	store.Close()
	var latestResultItems []models.LatestResult
	for _, late := range latest {
		latestResultItems = append(latestResultItems, models.LatestResult{
			Name:        late.Title,
			Description: late.Description,
			Site:        late.Site,
			Link:        late.SourceLink,
		})
	}
	return latestResultItems
}

func (st *StandardReportGenerator) GetIndexesStatus(indexFacade *Facade) []models.IndexStatus {
	var statuses []models.IndexStatus
	for ixKey, ix := range indexFacade.Scope.Indexes() {
		_, isAggregate := ix.(*Aggregate)
		statuses = append(statuses, models.IndexStatus{
			Index:       ixKey,
			IsAggregate: isAggregate,
			Errors:      ix.Errors(),
		})
	}
	return statuses
}
