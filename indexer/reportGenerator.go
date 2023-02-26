package indexer

import (
	config "github.com/sp0x/torrentd/config"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/indexer/status/models"
	"github.com/sp0x/torrentd/storage"
)

type StandardReportGenerator struct {
	config config.Config
}

//go:generate mockgen -source reportGenerator.go -destination=reportGeneratorMocks.go -package=indexer
type ReportGenerator interface {
	GetLatestItems() []models.LatestResult
	GetIndexesStatus(s *Facade) []models.IndexStatus
}

func NewStandardStatusReportGenerator(conf config.Config) ReportGenerator {
	return &StandardReportGenerator{
		config: conf,
	}
}

func (st *StandardReportGenerator) GetLatestItems() []models.LatestResult {
	strg := storage.NewBuilder(st.config).
		WithRecord(&search.TorrentResultItem{}).
		Build()
	latest := strg.GetLatest(20)
	strg.Close()
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
	store := storage.NewBuilder(st.config).
		WithRecord(&search.ScrapeResultItem{}).
		Build()
	storageStats := store.GetStats(false)
	store.Close()

	indexCount := len(indexFacade.IndexScope.Indexes())
	statuses := make([]models.IndexStatus, indexCount)
	statusIndex := 0
	for indexKey, indexes := range indexFacade.IndexScope.Indexes() {
		if indexes == nil {
			statusIndex += 1
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
		statuses[statusIndex] = indexStats
		statusIndex += 1
	}
	return statuses
}
