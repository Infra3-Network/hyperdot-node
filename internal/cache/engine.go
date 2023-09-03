package cache

import (
	"fmt"
	"sync"

	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

type EngineCache struct {
	tags map[string]map[string]*datamodel.QueryEngineDatasetInfo
	lock *sync.RWMutex
}

func NewDataEngineCache() *EngineCache {
	return &EngineCache{
		tags: make(map[string]map[string]*datamodel.QueryEngineDatasetInfo),
		lock: &sync.RWMutex{},
	}
}

var GlobalDataEngine = NewDataEngineCache()

func (e *EngineCache) SetDatasets(queryEngine string, datasets *datamodel.QueryEngineDatasets) {
	e.lock.Lock()
	defer e.lock.Unlock()
	if engine, ok := e.tags[queryEngine]; ok {
		if datasets.Raw != nil {
			engine["raw"] = datasets.Raw
		}
	}

}

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
