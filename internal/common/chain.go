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

var GlobalParaChainMap *ParaChainMap

func init() {
	GlobalParaChainMap = &ParaChainMap{
		data: make(map[int]ParaChainData),
		lock: new(sync.RWMutex),
	}
}
