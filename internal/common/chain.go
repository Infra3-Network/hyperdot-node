package common

import "sync"

type ParaChainData struct {
	ID              string `json:"id"`
	Prefix          int    `json:"prefix"`
	ChainID         int    `json:"chainID"`
	ChainName       string `json:"chainName"`
	Symbol          string `json:"symbol"`
	LastFinalizedTS int    `json:"lastFinalizedTS"`
	IconURL         string `json:"iconUrl"`
}

type ParaChainMap struct {
	data map[int]ParaChainData
	lock *sync.RWMutex
}

var GlobalParaChainCache *ParaChainMap

func NewParaChainMap() *ParaChainMap {
	return &ParaChainMap{
		data: make(map[int]ParaChainData),
		lock: new(sync.RWMutex),
	}
}

func (p *ParaChainMap) From(chains []ParaChainData) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, chain := range chains {
		p.data[chain.ChainID] = chain
	}
}

func (p *ParaChainMap) GetChains() map[int]ParaChainData {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.data
}
