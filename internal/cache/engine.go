package cache

import (
	"fmt"
	"sync"

	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

// EngineCache is a cache for query engine datasets.
// It is used to store the datasets of query engines and
// provide a way to get the datasets by query engine and tag.
type EngineCache struct {
	tags map[string]map[string]*datamodel.QueryEngineDatasetInfo
	lock *sync.RWMutex
}

// NewDataEngineCache creates a new EngineCache.
func NewDataEngineCache() *EngineCache {
	return &EngineCache{
		tags: make(map[string]map[string]*datamodel.QueryEngineDatasetInfo),
		lock: &sync.RWMutex{},
	}
}

// GlobalDataEngine is a global static EngineCache object.
var GlobalDataEngine = NewDataEngineCache()

// SetDatasets sets the datasets of a query engine.
func (e *EngineCache) SetDatasets(queryEngine string, datasets *datamodel.QueryEngineDatasets) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if engine, ok := e.tags[queryEngine]; ok {
		if datasets.Raw != nil {
			engine["raw"] = datasets.Raw
		}
	}

}

// GetDatasets gets the datasets of a query engine by tag.
func (e *EngineCache) GetDatasets(queryEngine string, tag string) (*datamodel.QueryEngineDatasetInfo, error) {
	e.lock.RLock()
	defer e.lock.RUnlock()
	if engine, ok := e.tags[queryEngine]; ok {
		if dataset, ok := engine[tag]; ok {
			return dataset, nil
		} else {
			return nil, fmt.Errorf("%s not found", tag)
		}
	} else {
		return nil, fmt.Errorf("%s not found", queryEngine)
	}
}
