package datamodel

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type ChainModel struct {
	ID                           string    `json:"id" gorm:"index:hyperdot_chains_idx_id"`
	Prefix                       int       `json:"prefix"`
	ChainID                      uint      `json:"chainID" gorm:"index:hyperdot_chains_idx_chain_id"`
	ChainName                    string    `json:"chainName"`
	Symbol                       string    `json:"symbol"`
	LastFinalizedTS              int       `json:"lastFinalizedTS"`
	IconURL                      string    `json:"iconUrl"`
	NumExtrinsics7d              int       `json:"numExtrinsics7d"`
	NumExtrinsics30d             int       `json:"numExtrinsics30d"`
	NumExtrinsics                int       `json:"numExtrinsics"`
	NumSignedExtrinsics7d        int       `json:"numSignedExtrinsics7d"`
	NumSignedExtrinsics30d       int       `json:"numSignedExtrinsics30d"`
	NumSignedExtrinsics          int       `json:"numSignedExtrinsics"`
	NumTransfers7d               int       `json:"numTransfers7d"`
	NumTransfers30d              int       `json:"numTransfers30d"`
	NumTransfers                 int       `json:"numTransfers"`
	NumEvents7d                  int       `json:"numEvents7d"`
	NumEvents30d                 int       `json:"numEvents30d"`
	NumEvents                    int       `json:"numEvents"`
	ValueTransfersUSD7d          float64   `json:"valueTransfersUSD7d"`
	ValueTransfersUSD30d         float64   `json:"valueTransfersUSD30d"`
	ValueTransfersUSD            float64   `json:"valueTransfersUSD"`
	NumXCMTransferIncoming       int       `json:"numXCMTransferIncoming"`
	NumXCMTransferIncoming7d     int       `json:"numXCMTransferIncoming7d"`
	NumXCMTransferIncoming30d    int       `json:"numXCMTransferIncoming30d"`
	NumXCMTransferOutgoing       int       `json:"numXCMTransferOutgoing"`
	NumXCMTransferOutgoing7d     int       `json:"numXCMTransferOutgoing7d"`
	NumXCMTransferOutgoing30d    int       `json:"numXCMTransferOutgoing30d"`
	ValXCMTransferIncomingUSD    float64   `json:"valXCMTransferIncomingUSD"`
	ValXCMTransferIncomingUSD7d  float64   `json:"valXCMTransferIncomingUSD7d"`
	ValXCMTransferIncomingUSD30d float64   `json:"valXCMTransferIncomingUSD30d"`
	ValXCMTransferOutgoingUSD    float64   `json:"valXCMTransferOutgoingUSD"`
	ValXCMTransferOutgoingUSD7d  float64   `json:"valXCMTransferOutgoingUSD7d"`
	ValXCMTransferOutgoingUSD30d float64   `json:"valXCMTransferOutgoingUSD30d"`
	NumTransactionsEVM           int       `json:"numTransactionsEVM"`
	NumTransactionsEVM7d         int       `json:"numTransactionsEVM7d"`
	NumTransactionsEVM30d        int       `json:"numTransactionsEVM30d"`
	NumHolders                   int       `json:"numHolders"`
	NumAccountsActive            int       `json:"numAccountsActive"`
	NumAccountsActive7d          int       `json:"numAccountsActive7d"`
	NumAccountsActive30d         int       `json:"numAccountsActive30d"`
	RelayChain                   string    `json:"relayChain"`
	TotalIssuance                int       `json:"totalIssuance"`
	IsEVM                        int       `json:"isEVM"`
	BlocksCovered                int       `json:"blocksCovered"`
	BlocksFinalized              int       `json:"blocksFinalized"`
	CrawlingStatus               string    `json:"crawlingStatus"`
	GithubURL                    string    `json:"githubURL"`
	SubstrateURL                 string    `json:"substrateURL"`
	ParachainsURL                string    `json:"parachainsURL"`
	DappURL                      string    `json:"dappURL"`
	Asset                        string    `json:"asset"`
	Decimals                     int       `json:"decimals"`
	PriceUSD                     float64   `json:"priceUSD"`
	PriceUSDPercentChange        float64   `json:"priceUSDPercentChange"`
	CreatedAt                    time.Time `json:"createdAt"`
	UpdatedAt                    time.Time `json:"updastedAt"`
	DeletedAt                    time.Time `json:"deleteAt"`
}

func (ChainModel) TableName() string {
	return "hyperdot_chains"
}

type RelayChainModel struct {
	ChainID    uint         `json:"chainID" gorm:"index:hyperdot_relay_chain_idx_chain_id"`
	SubChainId uint         `json:"subChainId" gorm:"index:hyperdot_relay_chain_idx_sub_chain_id"`
	Relay      ChainModel   `json:"relay"`
	Chains     []ChainModel `json:"chains"`
}

func HackAutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&ChainModel{}); err != nil {
		return err
	}

	return nil
}

type RelayChain struct {
	Relay  ChainModel   `json:"relay"`
	Chains []ChainModel `json:"chains"`
}

type BigQueryDataEngine struct {
	Name             string                 `json:"name"`
	Relays           []string               `json:"relays"`
	RelayChains      map[string]*RelayChain `json:"relayChains"`
	ChainTables      map[int][]Table        `json:"chainTables"` // chainid to tables
	CrossChainTables []Table                `json:"crossChainTables"`
	SystemTables     []Table                `json:"systemTables"`
}

func (b *BigQueryDataEngine) ToJSON() ([]byte, error) {
	return json.Marshal(b)
}
