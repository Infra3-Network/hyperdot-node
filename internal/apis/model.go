package apis

import "infra-3.xyz/hyperdot-node/internal/common"

type EngineModel struct {
	Name   string                       `json:"name"`
	Chains map[int]common.ParaChainData `json:"chains"`
	Tables []string                     `json:"tables"`
}
