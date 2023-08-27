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
	bigqueryClient *clients.SimpleBigQueryClinet
}

// New fetchpara chain
func NewBigQuerySyncer(cfg *common.Config) (*BigQuerySyncer, error) {
	ctx := context.Background()
	client, err := clients.NewSimpleBigQueryClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &BigQuerySyncer{
		ctx:            ctx,
		cfg:            *cfg,
		bigqueryClient: client,
	}, nil
}

func (f *BigQuerySyncer) Do() (*datamodel.BigQueryDataEngine, error) {
	return f.do()
}

func (f *BigQuerySyncer) do() (*datamodel.BigQueryDataEngine, error) {
	return BuildBigQueryEngine(f.ctx, f.bigqueryClient, &f.cfg.Polkaholic)
}

func BuildBigQueryEngine(ctx context.Context, bigqueryClient *clients.SimpleBigQueryClinet, cfg *common.PolkaholicConfig) (*datamodel.BigQueryDataEngine, error) {
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

	relayChainMap := make(map[string]*datamodel.RelayChain)
	// get relaychain
	for _, chain := range chains {
		if chain.ID == chain.RelayChain {
			relayChainMap[chain.RelayChain] = &datamodel.RelayChain{
				Relay:  chain,
				Chains: make([]datamodel.Chain, 0),
			}
		}
	}

	// get chains of relaychain
	for _, chain := range chains {
		if relayChain, ok := relayChainMap[chain.RelayChain]; ok && !(chain.ID == chain.RelayChain) {
			relayChain.Chains = append(relayChain.Chains, chain)
		}
	}

	chainTableMap := make(map[int][]datamodel.Table)
	crossChainTables := []datamodel.Table{}
	systemTables := []datamodel.Table{}

	tables, err := bigqueryClient.QueryCryptoPolkadotTableScheme(ctx)
	if err != nil {
		return nil, err
	}
	if err := processTables("polkadot", tables, chainTableMap, &crossChainTables, &systemTables); err != nil {
		return nil, err
	}

	tables, err = bigqueryClient.QueryCryptoKusamaTableScheme(ctx)
	if err != nil {
		return nil, err
	}
	if err := processTables("kusama", tables, chainTableMap, &crossChainTables, &systemTables); err != nil {
		return nil, err
	}

	var relays []string
	for relayChainName, _ := range relayChainMap {
		relays = append(relays, relayChainName)
	}

	return &datamodel.BigQueryDataEngine{
		Name:             "bigquery",
		Relays:           relays,
		RelayChains:      relayChainMap,
		ChainTables:      chainTableMap,
		CrossChainTables: crossChainTables,
		SystemTables:     systemTables,
	}, err
}

func processTables(relayChainName string, tables []datamodel.Table, chainTableMap map[int][]datamodel.Table, crossChainTables *[]datamodel.Table, systemTables *[]datamodel.Table) error {
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

		if tables, ok := chainTableMap[chainId]; ok {
			chainTableMap[chainId] = append(tables, table)
		} else {
			chainTableMap[chainId] = []datamodel.Table{table}
		}

	}
	return nil
}
