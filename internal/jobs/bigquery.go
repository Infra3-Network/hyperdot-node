package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"infra-3.xyz/hyperdot-node/internal/store"

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

// BigQuerySyncer is a job to sync bigquery engine chaindata
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

// Do executes the job
func (f *BigQuerySyncer) Do() error {
	chainData, err := f.do()
	if err != nil {
		log.Printf("Error fetching bigquery engine chaindata: %v", err)
		return err
	}

	// set raw
	if err := chainData.Raw.WriteToRedis(f.ctx, &f.cfg.Redis, "bigquery"); err != nil {
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

// BuildBigQueryEngineRawDataset creates a BigQuery dataset for the Polkadot and Kusama chains,
// populating information about chains, relay chains, and associated tables.
func BuildBigQueryEngineRawDataset(ctx context.Context, bigqueryClient *clients.SimpleBigQueryClient, cfg *common.PolkaholicConfig) (*datamodel.QueryEngineDatasetInfo, error) {
	// Log the start of the BuildBigQueryEngine job
	log.Printf("Start BuildBigQueryEngine Job")

	// Create a request to fetch chain information from the Polkaholic API
	url := fmt.Sprintf("%s/chains?limit=-1", cfg.BaseUrl)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Set the API key in the request header for authorization
	req.Header.Add("Authorization", cfg.ApiKey)

	// Execute the HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response into a slice of ChainModel
	var chains []datamodel.ChainModel
	err = json.NewDecoder(resp.Body).Decode(&chains)
	if err != nil {
		return nil, err
	}

	// Create a map to quickly access chain information by ChainID
	var chainMap = make(map[uint]datamodel.ChainModel, len(chains))
	for _, chain := range chains {
		chainMap[chain.ChainID] = chain
	}

	// Create a map to store metadata about relay chains
	relayChainMap := make(map[string]*datamodel.RelayChainMetadata, 0)
	for _, chain := range chains {
		// Check if the chain is the relay chain for itself
		if chain.ID == chain.RelayChain {
			var showColor string
			// Set color based on the chain name
			if chain.ChainName == "Polkadot" {
				showColor = "#E0016A"
			} else if chain.ChainName == "Kusama" {
				showColor = "#000000"
			} else {
				showColor = "#00C67D"
			}
			// Populate relay chain metadata
			relayChainMap[chain.RelayChain] = &datamodel.RelayChainMetadata{
				ChainID:      chain.ChainID,
				Name:         chain.ChainName,
				ShowColor:    showColor,
				ParaChainIDs: []uint{chain.ChainID},
			}
		}
	}

	// Populate para chain IDs for each relay chain
	for _, chain := range chains {
		if relayChain, ok := relayChainMap[chain.RelayChain]; ok && !(chain.ID == chain.RelayChain) {
			relayChain.ParaChainIDs = append(relayChain.ParaChainIDs, chain.ChainID)
		}
	}

	// Create maps to store tables for each chain and cross-chain tables
	chainTableMap := make(map[int][]datamodel.Table)
	crossChainTables := []datamodel.Table{}
	systemTables := []datamodel.Table{}

	// Query and process tables for the Polkadot chain
	tables, err := bigqueryClient.QueryCryptoPolkadotTableScheme(ctx)
	if err != nil {
		return nil, err
	}
	if err := processTables("polkadot", tables, &chainTableMap, &crossChainTables, &systemTables); err != nil {
		return nil, err
	}

	// Query and process tables for the Kusama chain
	kusamaTables, err := bigqueryClient.QueryCryptoKusamaTableScheme(ctx)
	if err != nil {
		return nil, err
	}

	if err := processTables("kusama", kusamaTables, &chainTableMap, &crossChainTables, &systemTables); err != nil {
		return nil, err
	}

	// Return the QueryEngineDatasetInfo containing the collected information
	return &datamodel.QueryEngineDatasetInfo{
		Id:          "raw",
		Chains:      chainMap,
		RelayChains: relayChainMap,
		ChainTables: chainTableMap,
	}, nil
}

// processTables populates tables based on relay chain name and categorizes them into cross-chain and system tables.
func processTables(relayChainName string, tables []datamodel.Table, chainTableMap *map[int][]datamodel.Table, crossChainTables *[]datamodel.Table, systemTables *[]datamodel.Table) error {
	// Regular expression to extract numeric chain ID from table names
	re := regexp.MustCompile(`\d+`)

	// Iterate over tables and categorize them
	for _, table := range tables {
		// Extract numeric chain ID from the table name
		match := re.FindString(table.TableID)

		// If no numeric chain ID is found
		if len(match) == 0 {
			// Check if the table is related to cross-chain transactions
			if strings.Contains(table.TableID, "xcm") {
				// Modify the table name and add it to cross-chain tables
				table.TableID = fmt.Sprintf("%s_%s", relayChainName, table.TableID)
				*crossChainTables = append(*crossChainTables, table)
			} else if table.TableID == "asserts" {
				// TODO: Handle asserts table
			} else if _, ok := systemTableMap[table.TableID]; ok {
				// Modify the table name and add it to system tables
				table.TableID = fmt.Sprintf("%s_%s", relayChainName, table.TableID)
				*systemTables = append(*systemTables, table)
			}
			continue
		}

		// Convert the numeric chain ID to an integer
		chainId, err := strconv.Atoi(match)
		if err != nil {
			return err
		}

		// Adjust chain ID for Kusama chain
		if relayChainName == "kusama" {
			// See https://github.com/colorfulnotion/substrate-etl/tree/main/kusama
			chainId += 20000
		}

		// Add the table to the corresponding chain in the chainTableMap
		if tables, ok := (*chainTableMap)[chainId]; ok {
			(*chainTableMap)[chainId] = append(tables, table)
		} else {
			(*chainTableMap)[chainId] = []datamodel.Table{table}
		}
	}
	return nil
}
