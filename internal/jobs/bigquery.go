package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"infra-3.xyz/hyperdot-node/internal/cache"
	"infra-3.xyz/hyperdot-node/internal/store"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

var (
	systemTableMap = map[string]struct{}{
		"AAA_tableschema": {},
		"chains":          {},
	}
)

type BigQuerySyncer struct {
	ctx            context.Context
	cfg            common.Config
	boltStore      *store.BoltStore
	bigqueryClient *clients.SimpleBigQueryClient
}

// NewBigQuerySyncer New fetchpara chain
func NewBigQuerySyncer(cfg *common.Config, boltStore *store.BoltStore) (*BigQuerySyncer, error) {
	ctx := context.Background()
	client, err := clients.NewSimpleBigQueryClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &BigQuerySyncer{
		ctx:            ctx,
		cfg:            *cfg,
		boltStore:      boltStore,
		bigqueryClient: client,
	}, nil
}

func (f *BigQuerySyncer) Do() error {
	chainData, err := f.do()
	if err != nil {
		log.Printf("Error fetching bigquery engine chaindata: %v", err)
		return err
	}

	cache.GlobalDataEngine.SetDatasets("bigquery", chainData) // TODO: should call SetDatasets
	if err := f.boltStore.SetDatasets("bigquery", chainData); err != nil {
		return err
	}

	return nil
}

func (f *BigQuerySyncer) do() (*datamodel.QueryEngineDatasets, error) {
	raw, err := BuildBigQueryEngineRawDataset(f.ctx, f.bigqueryClient, &f.cfg.Polkaholic)
	if err != nil {
		return nil, err
	}

	return &datamodel.QueryEngineDatasets{
		Raw: raw,
	}, nil
}

func BuildBigQueryEngineRawDataset(ctx context.Context, bigqueryClient *clients.SimpleBigQueryClient, cfg *common.PolkaholicConfig) (*datamodel.QueryEngineDatasetInfo, error) {
	log.Printf("Start BuildBigQueryEngine Job")

	url := fmt.Sprintf("%s/chains?limit=-1", cfg.BaseUrl)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", cfg.ApiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error making GET request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	var chains []datamodel.Chain
	err = json.NewDecoder(resp.Body).Decode(&chains)
	if err != nil {
		fmt.Println("Error decoding JSON response:", err)
		return nil, err
	}

	var chainMap = make(map[int]datamodel.Chain, len(chains))
	for _, chain := range chains {
		chainMap[chain.ChainID] = chain
	}

	relayChainMap := make(map[string]*datamodel.RelayChainMetadata, 0)
	for _, chain := range chains {
		if chain.ID == chain.RelayChain {
			var showColor string
			if chain.ChainName == "Polkadot" {
				showColor = "#E0016A"
			} else if chain.ChainName == "Kusama" {
				showColor = "#000000"
			} else {
				showColor = "#00C67D"
			}
			relayChainMap[chain.RelayChain] = &datamodel.RelayChainMetadata{
				ChainID:   chain.ChainID,
				Name:      chain.ChainName,
				ShowColor: showColor,
				ParaChainIDs: []int{
					chain.ChainID,
				},
			}
		}
	}
	//relayChainMap := make(map[string]*datamodel.RelayChain)
	//for _, chain := range chains {
	//	if chain.ID == chain.RelayChain {
	//		relayChainMap[chain.RelayChain] = &datamodel.RelayChain{
	//			Relay:  chain,
	//			Chains: make([]datamodel.Chain, 0),
	//		}
	//	}
	//}

	// get chains of relaychain
	for _, chain := range chains {
		if relayChain, ok := relayChainMap[chain.RelayChain]; ok && !(chain.ID == chain.RelayChain) {
			relayChain.ParaChainIDs = append(relayChain.ParaChainIDs, chain.ChainID)
		}
	}

	chainTableMap := make(map[int][]datamodel.Table)
	crossChainTables := []datamodel.Table{}
	systemTables := []datamodel.Table{}

	tables, err := bigqueryClient.QueryCryptoPolkadotTableScheme(ctx)
	if err != nil {
		return nil, err
	}
	if err := processTables("polkadot", tables, &chainTableMap, &crossChainTables, &systemTables); err != nil {
		return nil, err
	}

	kusamaTables, err := bigqueryClient.QueryCryptoKusamaTableScheme(ctx)
	if err != nil {
		return nil, err
	}

	if err := processTables("kusama", kusamaTables, &chainTableMap, &crossChainTables, &systemTables); err != nil {
		return nil, err
	}

	var relays []string
	for relayChainName, _ := range relayChainMap {
		relays = append(relays, relayChainName)
	}

	return &datamodel.QueryEngineDatasetInfo{
		Id:          "raw",
		Chains:      chainMap,
		RelayChains: relayChainMap,
		ChainTables: chainTableMap,
	}, err
}

func processTables(relayChainName string, tables []datamodel.Table, chainTableMap *map[int][]datamodel.Table, crossChainTables *[]datamodel.Table, systemTables *[]datamodel.Table) error {
	re := regexp.MustCompile(`\d+`)
	for _, table := range tables {
		match := re.FindString(table.TableID)
		if len(match) == 0 {
			if strings.Contains(table.TableID, "xcm") {
				table.TableID = fmt.Sprintf("%s_%s", relayChainName, table.TableID)
				*crossChainTables = append(*crossChainTables, table)
			} else if table.TableID == "asserts" {
				// TODO:
			} else if _, ok := systemTableMap[table.TableID]; ok {
				table.TableID = fmt.Sprintf("%s_%s", relayChainName, table.TableID)
				*systemTables = append(*systemTables, table)
			}
			continue
		}

		chainId, err := strconv.Atoi(match)
		if err != nil {
			return err
		}

		if relayChainName == "kusama" {
			// see https://github.com/colorfulnotion/substrate-etl/tree/main/kusama
			chainId += 20000
		}

		if tables, ok := (*chainTableMap)[chainId]; ok {
			(*chainTableMap)[chainId] = append(tables, table)
		} else {
			(*chainTableMap)[chainId] = []datamodel.Table{table}
		}
	}
	fmt.Println("chainTableMapLen: ", len(*chainTableMap))
	return nil
}
