package indexer

import (
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/status/models"
	"github.com/sp0x/torrentd/storage"
)

type StandardReportGenerator struct{}

//go:generate mockgen -source reportGenerator.go -destination=reportGeneratorMocks.go -package=indexer
type ReportGenerator interface {
	GetLatestItems() []models.LatestResult
	GetIndexesStatus(s *Facade) []models.IndexStatus
}

func (st *StandardReportGenerator) GetLatestItems() []models.LatestResult {
	store := storage.NewBuilder().
		WithRecord(&search.TorrentResultItem{}).
		Build()
	latest := store.GetLatest(20)
	store.Close()
	latestResultItems := make([]models.LatestResult, len(latest))
	for _, late := range latest {
		torrentItem := late.(*search.TorrentResultItem)
		latestResultItems = append(latestResultItems, models.LatestResult{
			Name:        torrentItem.Title,
			Description: torrentItem.Description,
			Site:        torrentItem.Site,
			Link:        torrentItem.SourceLink,
		})
	}
	return latestResultItems
}

func (st *StandardReportGenerator) GetIndexesStatus(indexFacade *Facade) []models.IndexStatus {
	store := storage.NewBuilder().
		WithRecord(&search.ScrapeResultItem{}).
		Build()
	storageStats := store.GetStats(false)
	store.Close()

	indexCount := len(indexFacade.IndexScope.Indexes())
	statuses := make([]models.IndexStatus, indexCount)

	for indexKey, indexes := range indexFacade.IndexScope.Indexes() {
		if indexes == nil {
			continue
		}
		isAggregate := len(indexes) > 1
		indexStats := models.IndexStatus{
			Index:       indexKey,
			IsAggregate: isAggregate,
			Errors:      indexes.Errors(),
		}
		if storageStats != nil {
			nsp := storageStats.GetNamespace(indexKey)
			if nsp != nil {
				indexStats.Size = nsp.RecordCount
			}
		}
		statuses = append(statuses, indexStats)
	}
	return statuses
}
