package chains

import "infra-3.xyz/hyperdot-node/internal/chains"

type Chain struct {
	ID                           string  `json:"id"`
	Prefix                       int     `json:"prefix"`
	ChainID                      int     `json:"chainID"`
	ChainName                    string  `json:"chainName"`
	Symbol                       string  `json:"symbol"`
	LastFinalizedTS              int     `json:"lastFinalizedTS"`
	IconURL                      string  `json:"iconUrl"`
	NumExtrinsics7d              int     `json:"numExtrinsics7d"`
	NumExtrinsics30d             int     `json:"numExtrinsics30d"`
	NumExtrinsics                int     `json:"numExtrinsics"`
	NumSignedExtrinsics7d        int     `json:"numSignedExtrinsics7d"`
	NumSignedExtrinsics30d       int     `json:"numSignedExtrinsics30d"`
	NumSignedExtrinsics          int     `json:"numSignedExtrinsics"`
	NumTransfers7d               int     `json:"numTransfers7d"`
	NumTransfers30d              int     `json:"numTransfers30d"`
	NumTransfers                 int     `json:"numTransfers"`
	NumEvents7d                  int     `json:"numEvents7d"`
	NumEvents30d                 int     `json:"numEvents30d"`
	NumEvents                    int     `json:"numEvents"`
	ValueTransfersUSD7d          float64 `json:"valueTransfersUSD7d"`
	ValueTransfersUSD30d         float64 `json:"valueTransfersUSD30d"`
	ValueTransfersUSD            float64 `json:"valueTransfersUSD"`
	NumXCMTransferIncoming       int     `json:"numXCMTransferIncoming"`
	NumXCMTransferIncoming7d     int     `json:"numXCMTransferIncoming7d"`
	NumXCMTransferIncoming30d    int     `json:"numXCMTransferIncoming30d"`
	NumXCMTransferOutgoing       int     `json:"numXCMTransferOutgoing"`
	NumXCMTransferOutgoing7d     int     `json:"numXCMTransferOutgoing7d"`
	NumXCMTransferOutgoing30d    int     `json:"numXCMTransferOutgoing30d"`
	ValXCMTransferIncomingUSD    float64 `json:"valXCMTransferIncomingUSD"`
	ValXCMTransferIncomingUSD7d  float64 `json:"valXCMTransferIncomingUSD7d"`
	ValXCMTransferIncomingUSD30d float64 `json:"valXCMTransferIncomingUSD30d"`
	ValXCMTransferOutgoingUSD    float64 `json:"valXCMTransferOutgoingUSD"`
	ValXCMTransferOutgoingUSD7d  float64 `json:"valXCMTransferOutgoingUSD7d"`
	ValXCMTransferOutgoingUSD30d float64 `json:"valXCMTransferOutgoingUSD30d"`
	NumTransactionsEVM           int     `json:"numTransactionsEVM"`
	NumTransactionsEVM7d         int     `json:"numTransactionsEVM7d"`
	NumTransactionsEVM30d        int     `json:"numTransactionsEVM30d"`
	NumHolders                   int     `json:"numHolders"`
	NumAccountsActive            int     `json:"numAccountsActive"`
	NumAccountsActive7d          int     `json:"numAccountsActive7d"`
	NumAccountsActive30d         int     `json:"numAccountsActive30d"`
	RelayChain                   string  `json:"relayChain"`
	TotalIssuance                int     `json:"totalIssuance"`
	IsEVM                        int     `json:"isEVM"`
	BlocksCovered                int     `json:"blocksCovered"`
	BlocksFinalized              int     `json:"blocksFinalized"`
	CrawlingStatus               string  `json:"crawlingStatus"`
	GithubURL                    string  `json:"githubURL"`
	SubstrateURL                 string  `json:"substrateURL"`
	ParachainsURL                string  `json:"parachainsURL"`
	DappURL                      string  `json:"dappURL"`
	Asset                        string  `json:"asset"`
	Decimals                     int     `json:"decimals"`
	PriceUSD                     float64 `json:"priceUSD"`
	PriceUSDPercentChange        float64 `json:"priceUSDPercentChange"`
}

type RelayChain struct {
	ChainID      int    `json:"chain_id"`
	Name         string `json:"name"`
	ShowColor    string `json:"show_color"`
	ParaChainIDs []int  `json:"para_chain_ids"`
}

type Dataset struct {
	Chains      map[int]Chain          `json:"chains"`
	RelayChains map[string]*RelayChain `json:"relay_chains"`
	ChainTables map[int][]chains.Table `json:"chain_tables"`
}
