package dashboard

import (
	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

// Response is the response struct for dashboard restful api
type Response struct {
	base.BaseResponse
	Data datamodel.DashboardModel `json:"data"`
}
