package indexer

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/sp0x/torrentd/indexer/search"
	"github.com/sp0x/torrentd/storage"
)

type indexWorkerPool struct {
	storage             storage.ItemStorage
	completionWaitGroup sync.WaitGroup
	iterators           map[Indexer]*search.SearchStateIterator
	query               *search.Query
	workChannel         chan *workerJob
	resultsChannel      chan []search.ResultItemBase
}

func newWorkerJob(iterator *search.SearchStateIterator, index Indexer, fields map[string]interface{}, page uint) *workerJob {
	return &workerJob{
		Iterator: iterator,
		Fields:   fields,
		Page:     page,
		id:       "",
		Index:    index,
	}
}

// feedWorkerPool Iterate over the index search iterators and add the data to the work channel
func (f *Facade) feedWorkerPool(workerPool *indexWorkerPool) {
	for !workerPool.isComplete() {
		for indexForIterator, iterator := range workerPool.iterators {
			if iterator.IsComplete() {
				continue
			}

			fields, page := iterator.Next()
			nextJob := newWorkerJob(iterator, indexForIterator, fields, page)
			f.logger.Debugf("Adding job %v to work channel", nextJob)
			workerPool.workChannel <- nextJob
			if iterator.IsComplete() {
				f.logger.Debugf("Completed iterator %p for index %v", iterator, indexForIterator.GetDefinition().Name)
			}
		}
	}
	close(workerPool.workChannel)
}

func (p *indexWorkerPool) isComplete() bool {
	if len(p.iterators) == 0 {
		return true
	}
	totalItemsDiscovered := uint(0)
	copletedAllIterators := true

	for _, iterator := range p.iterators {
		totalItemsDiscovered += iterator.GetItemsDiscoveredCount()
		if p.query.HasEnoughResults(totalItemsDiscovered) {
			return true
		}

		if copletedAllIterators && !iterator.IsComplete() {
			copletedAllIterators = false
		}
	}

	return copletedAllIterators
}

//region Workers

func (f *Facade) runWorker(
	id int,
	query *search.Query,
	wg *sync.WaitGroup,
	resultStorage storage.ItemStorage,
	workChannel <-chan *workerJob,
	resultsChannel chan<- []search.ResultItemBase) {

	for workJob := range workChannel {
		log.Debugf("Gor work job: %v", workJob)
		searchResults, err := workJob.Index.Search(query, workJob)
		if err != nil {
			log.WithFields(log.Fields{}).Errorf("Couldn't persist item: %s\n", err)
			continue
		}

		if saveResultsOnDiscovery {
			saveDiscoveredItems(searchResults, resultStorage)
		}

		workJob.Iterator.UpdateIteratorState(searchResults)

		if searchResults != nil {
			resultsChannel <- searchResults
		}
		if workJob.Iterator.IsComplete() {
			break
		}
	}

	wg.Done()
	f.logger.Debugf("Worker #%v ran out of jobs", id)
}

func saveDiscoveredItems(searchResults []search.ResultItemBase, resultStorage storage.ItemStorage) {
	for _, item := range searchResults {
		err := resultStorage.Add(item)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).
				Error("Couldn't persist item.")
		}
	}
}

/// createWorkerPool Creates a pool of workers that run in the background using work and results channels
func (f Facade) createWorkerPool(indexes []Indexer, resultStorage storage.ItemStorage, query *search.Query, workerCount int) *indexWorkerPool {
	workerPool := &indexWorkerPool{}
	workerPool.workChannel = make(chan *workerJob, workerCount)
	workerPool.resultsChannel = make(chan []search.ResultItemBase, workerCount)
	workerPool.storage = resultStorage
	workerPool.iterators = make(map[Indexer]*search.SearchStateIterator)
	workerPool.query = query

	for w := 0; w < workerCount; w++ {
		workerPool.completionWaitGroup.Add(1)
		go f.runWorker(w, query, &workerPool.completionWaitGroup, resultStorage, workerPool.workChannel, workerPool.resultsChannel)
	}

	for _, index := range indexes {
		workerPool.iterators[index] = search.NewIterator(query)
	}

	go func() {
		// Wait for pool to be complete and close the results channel
		workerPool.completionWaitGroup.Wait()
		close(workerPool.resultsChannel)
	}()

	return workerPool
}

//endregion

//region Worker job

func (s workerJob) String() string {
	output := make([]string, len(s.Fields)+1)
	i := 0
	output[0] = fmt.Sprintf("page: %d", s.Page)
	for fname, fval := range s.Fields {
		val := fmt.Sprintf("{%s: %v}", fname, fval)
		output[i+1] = val
		i++
	}
	return strings.Join(output, ",")
}

func (s *workerJob) SetID(id string) {
	s.id = id
}

//endregion
