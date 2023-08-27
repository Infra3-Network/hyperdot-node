package cache

import (
	"sync"

	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

type engineCache struct {
	bigquery *datamodel.BigQueryDataEngine
	lock     *sync.RWMutex
}

func NewDataEngineCache() *engineCache {
	return &engineCache{
		bigquery: new(datamodel.BigQueryDataEngine),
		lock:     &sync.RWMutex{},
	}
}

var GlobalDataEngine = NewDataEngineCache()

func (e *engineCache) SetBigQuery(engine *datamodel.BigQueryDataEngine) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.bigquery = engine
}

func (e *engineCache) GetBigQuery() *datamodel.BigQueryDataEngine {
	e.lock.RLock()
	defer e.lock.RUnlock()

	return e.bigquery
}
