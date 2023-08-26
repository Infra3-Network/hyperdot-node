package apis

import "github.com/gin-gonic/gin"

type QueryService struct {
	group string
}

func NewQueryService() *QueryService {
	return &QueryService{
		group: "/query",
	}
}

func (q *QueryService) ListEnginesHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(200, ListEngineResponse{
			BaseResponse: ResponseOk(),
			Engines: []EngineModel{
				{
					Name: "bigquery",
				},
			},
		})
	}
}

func (q *QueryService) Name() string {
	return "query"
}

func (q *QueryService) RouteTables() []RouteTable {
	return []RouteTable{
		{
			Method:  "GET",
			Group:   q.group,
			Path:    q.group + "/engines",
			Handler: q.ListEnginesHandle(),
		},
	}
}
