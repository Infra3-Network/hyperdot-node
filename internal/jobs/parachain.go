package jobs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"infra-3.xyz/hyperdot-node/internal/common"
)

type ParachainData struct {
	ID              string `json:"id"`
	Prefix          int    `json:"prefix"`
	ChainID         int    `json:"chainID"`
	ChainName       string `json:"chainName"`
	Symbol          string `json:"symbol"`
	LastFinalizedTS int    `json:"lastFinalizedTS"`
	IconURL         string `json:"iconUrl"`
}

type FetchParaChain struct {
	cfg common.Config
}

// New fetchpara chain
func NewFetchParaChain(cfg *common.Config) *FetchParaChain {
	return &FetchParaChain{
		cfg: *cfg,
	}
}

func (f *FetchParaChain) Run() {
	ticker := time.Tick(time.Hour)
	for {
		<-ticker
		f.do()
	}
}

func (f *FetchParaChain) Do() ([]common.ParaChainData, error) {
	return f.do()
}

func (f *FetchParaChain) do() ([]common.ParaChainData, error) {
	url := fmt.Sprintf("%s/chains?limit=-1", f.cfg.Pokaholic.BaseUrl)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", f.cfg.Pokaholic.ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var chains []common.ParaChainData
	err = json.NewDecoder(resp.Body).Decode(&chains)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return nil, err
	}

	return chains, err
}
