package query

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
	"infra-3.xyz/hyperdot-node/internal/dataengine"

	"cloud.google.com/go/bigquery"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/iterator"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/cache"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
	"infra-3.xyz/hyperdot-node/internal/store"
)

const Name = "query"

type Service struct {
	group          string
	db             *gorm.DB
	bboltStore     *store.BoltStore
	bigqueryClient *clients.SimpleBigQueryClient
	engines        map[string]dataengine.QueryEngine
}

func New(bboltStore *store.BoltStore, cfg *common.Config, db *gorm.DB, engines map[string]dataengine.QueryEngine) *Service {
	ctx := context.Background()
	bigqueryClient, err := clients.NewSimpleBigQueryClient(ctx, cfg)
	if err != nil {
		panic(err)
	}
	return &Service{
		group:          "/query",
		db:             db,
		bboltStore:     bboltStore,
		bigqueryClient: bigqueryClient,
		engines:        engines,
	}
}

func (s *Service) ListEnginesHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get query engines from bolt
		data, err := s.bboltStore.GetQueryEngines()
		if err != nil {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: err.Error(),
			})
			return
		}

		ctx.JSON(200, base.ListEngineResponse{
			BaseResponse: base.ResponseOk(),
			Data:         data,
		})
	}
}

func (s *Service) GetQueryEngineDatasetHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get :engineId
		engineId := ctx.Param("engineId")
		if len(engineId) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "engineId is empty",
			})
			return
		}

		// get :datasetId
		datasetId := ctx.Param("datasetId")
		if len(datasetId) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "datasetId is empty",
			})
			return
		}

		engineId = strings.ToLower(engineId)
		datasetId = strings.ToLower(datasetId)

		var err error
		var data *datamodel.QueryEngineDatasetInfo
		data, err = cache.GlobalDataEngine.GetDatasets(engineId, datasetId)
		if err != nil || data == nil {
			data, err = s.bboltStore.GetDataset(engineId, datasetId)
			if err != nil {
				ctx.JSON(200, base.BaseResponse{
					ErrorCode:    base.Err,
					ErrorMessage: err.Error(),
				})
				return
			}
		}

		chainTables := make(map[int][]string)
		for chainID, tables := range data.ChainTables {
			for _, table := range tables {
				chainTables[chainID] = append(chainTables[chainID], table.TableID)
			}
		}
		ctx.JSON(200, base.GetQueryEngineDatasetResponse{
			BaseResponse: base.ResponseOk(),
			Data: struct {
				Id          string                                   `json:"id"`
				Chains      map[int]datamodel.Chain                  `json:"chains"`
				RelayChains map[string]*datamodel.RelayChainMetadata `json:"relayChains"`
				ChainTables map[int][]string                         `json:"chainTables"`
			}(struct {
				Id          string
				Chains      map[int]datamodel.Chain
				RelayChains map[string]*datamodel.RelayChainMetadata
				ChainTables map[int][]string
			}{Id: data.Id, Chains: data.Chains, RelayChains: data.RelayChains, ChainTables: chainTables}),
		})
	}
}

// @BasePath /apis/v1/

// @Summary run query
// @Schemes
// @Description run query
// @Accept json
// @Produce json
// @Success 200 {QueryRunResponseData} QueryRunResponseData
// @Router /apis/v1/query/run [post]
func (s *Service) RunHandle() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// extract QueryRunRequest
		var req base.QueryRunRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: err.Error(),
			})
			return
		}

		if len(req.Query) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "Query is empty",
			})
			return
		}

		if len(req.Engine) == 0 {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: "Engine is empty",
			})
			return
		}

		if strings.ToLower(req.Engine) != "bigquery" {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: fmt.Sprintf("Engine %s is not supported", req.Engine),
			})
			return
		}

		// run query
		iter, err := s.bigqueryClient.Query(ctx, req.Query)
		if err != nil {
			ctx.JSON(200, base.BaseResponse{
				ErrorCode:    base.Err,
				ErrorMessage: err.Error(),
			})
			return
		}

		var schemas []datamodel.TableSchema
		for _, filed := range iter.Schema {
			mode := ""
			if filed.Repeated {
				mode = "REPEATED"
			} else if filed.Required {
				mode = "REQUIRED"
			} else {
				mode = "NULLABLE"
			}

			schemas = append(schemas, datamodel.TableSchema{
				Mode: mode,
				Name: filed.Name,
				Type: string(filed.Type),
			})
		}

		// extract result
		var results []map[string]bigquery.Value
		for {
			var row map[string]bigquery.Value
			err := iter.Next(&row)
			if errors.Is(err, iterator.Done) {
				break
			} else if err != nil {
				log.Fatalf("Error iterating through results: %v", err)
			}
			results = append(results, row)
		}

		ctx.JSON(200, base.QueryRunResponse{
			BaseResponse: base.ResponseOk(),
			Data:         base.QueryRunResponseData{Schemas: schemas, Rows: results},
		})
	}
}

func (s *Service) runHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		query := ctx.Query("q")
		engineName := ctx.Query("engine")
		_, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if len(query) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "query engine is required")
			return
		}
		if len(engineName) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "query is required")
			return
		}

		engine, ok := s.engines[engineName]
		if !ok {
			base.ResponseErr(ctx, http.StatusBadRequest, "The %s query engine unsupported now", engineName)
			return
		}

		iter, err := engine.Run(context.Background(), query)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, "query error: %v", err)
			return
		}

		var rows []map[string]interface{}
		for {
			row, err := iter.Next()
			if err != nil {
				if errors.Is(err, dataengine.IterDone) {
					break
				}
				base.ResponseErr(ctx, http.StatusBadRequest, "query error: %v", err)
				return
			}
			rows = append(rows, row)
		}

		schemas := iter.Schema()

		ctx.JSON(http.StatusOK, RunResponse{
			BaseResponse: base.BaseResponse{
				Success: true,
			},
			Data: RunResponseData{
				Rows:    rows,
				Schemas: schemas,
			},
		})
	}
}

func (s *Service) checkUserQueryModelRequest(ctx *gin.Context, model *datamodel.QueryModel) bool {
	currentUserId, err := base.GetCurrentUserId(ctx)
	if err != nil {
		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		return false
	}

	if model.UserID != 0 && model.UserID != currentUserId {
		base.ResponseErr(ctx, http.StatusUnauthorized, "Unauthorized")
		return false
	}

	model.UserID = currentUserId

	if len(model.QueryEngine) == 0 {
		base.ResponseErr(ctx, http.StatusBadRequest, "query engine is required")
		return false
	}
	if len(model.Query) == 0 {
		base.ResponseErr(ctx, http.StatusBadRequest, "query is required")
		return false
	}

	if !model.Unsaved {
		if len(model.Name) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "name is required")
			return false
		}
	} else {
		model.Name = "unsaved"
	}

	return true
}

func (s *Service) getQueryHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var query datamodel.QueryModel
		result := s.db.First(&query, id)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				base.ResponseErr(ctx, http.StatusOK, "query not found")
				return
			}
			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		if err := s.db.Where("query_id = ? AND user_id = ?", query.ID, query.UserID).Find(&query.Charts).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, Response{
			BaseResponse: base.BaseResponse{
				Success: true,
			},
			Data: query,
		})

	}
}

func (s *Service) getListParams(ctx *gin.Context) (*prePareListSQLParams, error) {
	var (
		err      error
		page     uint
		pageSize uint
		userId   uint
	)
	page, err = base.GetUIntQuery(ctx, "page")
	if err != nil {
		if err == base.ErrQueryNotFound {
			page = 1
		} else {
			return nil, err
		}
	}

	pageSize, err = base.GetUIntQuery(ctx, "page_size")
	if err != nil {
		if err == base.ErrQueryNotFound {
			pageSize = 10
		} else {
			return nil, err
		}
	}

	userId, err = base.GetUIntQuery(ctx, "user_id")
	if err != nil {
		if err == base.ErrQueryNotFound {
			userId = 0
		} else {
			return nil, err
		}
	}

	timeRange, err := base.GetStringQuery(ctx, "time_range")
	if err != nil {
		if err == base.ErrQueryNotFound {
			timeRange = "all"
		} else {
			return nil, err
		}
	}

	order, err := base.GetStringQuery(ctx, "order")
	if err != nil {
		if err == base.ErrQueryNotFound {
			order = "trending"
		} else {
			return nil, err
		}
	}

	return &prePareListSQLParams{
		Page:      page,
		PageSize:  pageSize,
		Order:     order,
		UserID:    userId,
		TimeRange: timeRange,
	}, nil
}

func (s *Service) listQueryHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		params, err := s.getListParams(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}
		params.CurrentUserId = currentUserId

		queryRaw, countRaw, err := s.prepareListSQL(params)
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		rows, err := queryRaw.Rows()
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		defer rows.Close()

		var queries []map[string]interface{}
		for rows.Next() {
			data := make(map[string]interface{})
			if err := s.db.ScanRows(rows, &data); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			queries = append(queries, data)
		}

		var total uint
		if rows, err = countRaw.Rows(); err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		defer rows.Close()

		for rows.Next() {
			if err := s.db.ScanRows(rows, &total); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			} else {
				break
			}
		}

		base.ResponseWithMap(ctx, map[string]interface{}{
			"queries": queries,
			"total":   total,
		})

		// currentUserId, err := base.GetCurrentUserId(ctx)
		// if err != nil {
		// 	base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		// 	return
		// }

		// var (
		// 	page     uint
		// 	pageSize uint
		// 	userId   uint
		// )

		// if page, err = base.GetUIntQuery(ctx, "page"); err != nil {
		// 	if err == base.ErrQueryNotFound {
		// 		page = 1
		// 	} else {
		// 		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		// 		return
		// 	}
		// }

		// if pageSize, err = base.GetUIntQuery(ctx, "page_size"); err != nil {
		// 	if err == base.ErrQueryNotFound {
		// 		pageSize = 10
		// 	} else {
		// 		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		// 	}
		// }

		// if userId, err = base.GetUIntQuery(ctx, "user_id"); err != nil && err != base.ErrQueryNotFound {
		// 	base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		// 	return
		// }

		// var raw *gorm.DB
		// tb1 := datamodel.QueryModel{}.TableName()
		// tb2 := datamodel.UserModel{}.TableName()
		// tb3 := datamodel.UserQueryFavorites{}.TableName()
		// if userId == 0 {
		// 	sql := `
		// 	SELECT
		// 		tb1.*,
		// 		tb2.username,
		// 		tb2.email,
		// 		tb2.uid,
		// 		tb2.icon_url,
		// 		tb3.stared
		// 	FROM
		// 		%s AS tb1
		// 		LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
		// 		LEFT JOIN %s AS tb3 ON tb1.id = tb3.query_id AND tb3.user_id = ?
		// 	WHERE
		// 		tb1.is_privacy = FALSE
		// 	ORDER BY
		// 		updated_at DESC
		// 		LIMIT ? OFFSET ( ? - 1 ) * ?
		// 	`
		// 	sql = fmt.Sprintf(sql, tb1, tb2, tb3)
		// 	raw = s.db.Raw(sql, currentUserId, pageSize, page, pageSize)
		// } else {
		// 	sql := `
		// 	SELECT
		// 		tb1.*,
		// 		tb2.username,
		// 		tb2.email,
		// 		tb2.uid,
		// 		tb2.icon_url,
		// 		tb3.stared
		// 	FROM
		// 		%s AS tb1
		// 		LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
		// 		LEFT JOIN %s AS tb3 ON tb1.id = tb3.query_id AND tb3.user_id = ?
		// 	WHERE
		// 		tb1.is_privacy = FALSE
		// 		AND tb1.user_id = ?
		// 	ORDER BY
		// 		updated_at DESC
		// 		LIMIT ? OFFSET ( ? - 1 ) * ?
		// 	`
		// 	sql = fmt.Sprintf(sql, tb1, tb2, tb3)
		// 	raw = s.db.Raw(sql, currentUserId, userId, pageSize, page, pageSize)
		// }

		// rows, err := raw.Rows()
		// if err != nil {
		// 	base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		// 	return
		// }
		// defer rows.Close()

		// var queries []map[string]interface{}
		// for rows.Next() {
		// 	data := make(map[string]interface{})
		// 	if err := s.db.ScanRows(rows, &data); err != nil {
		// 		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		// 		return
		// 	}
		// 	// convert chart string to structured chart
		// 	rawChart, ok := data["charts"]
		// 	if ok && rawChart != nil {
		// 		jsonChart := make([]map[string]interface{}, 0)
		// 		if err := json.Unmarshal([]byte(rawChart.(string)), &jsonChart); err != nil {
		// 			log.Printf("unmarshal chart error: %v", err)
		// 			continue
		// 		}
		// 		data["charts"] = jsonChart
		// 	}

		// 	var stars int64
		// 	if err = s.db.Table(tb3).Where("query_id = ? and stared = true", data["id"]).Count(&stars).Error; err != nil {
		// 		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		// 		return
		// 	}
		// 	data["stars"] = stars

		// 	queries = append(queries, data)
		// }

		// // get total
		// if userId == 0 {
		// 	sql := `
		// 	SELECT COUNT
		// 		( ID )
		// 	FROM
		// 		%s
		// 	WHERE
		// 		is_privacy = FALSE
		// 	`
		// 	sql = fmt.Sprintf(sql, tb1)
		// 	raw = s.db.Raw(sql)
		// } else {
		// 	sql := `
		// 	SELECT COUNT
		// 		( ID )
		// 	FROM
		// 		%s
		// 	WHERE
		// 		is_privacy = FALSE
		// 		AND user_id = ?
		// 	`
		// 	sql = fmt.Sprintf(sql, tb1)
		// 	raw = s.db.Raw(sql, userId)
		// }

		// var total uint
		// if rows, err = raw.Rows(); err != nil {
		// 	base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		// 	return
		// }
		// defer rows.Close()

		// for rows.Next() {
		// 	if err := s.db.ScanRows(rows, &total); err != nil {
		// 		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		// 		return
		// 	} else {
		// 		break
		// 	}
		// }

		// base.ResponseWithMap(ctx, map[string]any{
		// 	"queries": queries,
		// 	"total":   total,
		// })

	}
}

func (s *Service) listFavoriteQueryHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		params, err := s.getListParams(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}
		params.CurrentUserId = currentUserId

		queryRaw, countRaw, err := s.prepareListStaredSQL(params)
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		rows, err := queryRaw.Rows()
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		defer rows.Close()

		var queries []map[string]interface{}
		for rows.Next() {
			data := make(map[string]interface{})
			if err := s.db.ScanRows(rows, &data); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			queries = append(queries, data)
		}

		var total uint

		if rows, err = countRaw.Rows(); err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		defer rows.Close()

		for rows.Next() {
			if err := s.db.ScanRows(rows, &total); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			} else {
				break
			}
		}

		base.ResponseWithMap(ctx, map[string]interface{}{
			"queries": queries,
			"total":   total,
		})
	}
}

func (s *Service) listCurrentUserQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		s.listUserQueryChart(ctx, userId)
	}
}

func (s *Service) listUserQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userId, err := base.GetUintParam(ctx, "userId")
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		s.listUserQueryChart(ctx, userId)
	}
}

func (s *Service) listUserQueryChart(ctx *gin.Context, userId uint) {
	var (
		err      error
		page     uint
		pageSize uint
	)

	if page, err = base.GetUIntQuery(ctx, "page"); err != nil {
		if err == base.ErrQueryNotFound {
			page = 1
		} else {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}

	if pageSize, err = base.GetUIntQuery(ctx, "page_size"); err != nil {
		if err == base.ErrQueryNotFound {
			pageSize = 10
		} else {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}

	var charts []map[string]any
	tb1 := datamodel.ChartModel{}.TableName()
	tb2 := datamodel.QueryModel{}.TableName()
	sql := `
		SELECT
			tb2.name AS query_name,
			tb2.description AS query_description,
			tb2.query,
			tb2.query_engine,
			tb2.is_privacy,
			tb2.unsaved,
			tb2.stars as query_stars,
			tb2.created_at AS query_created_at,
			tb2.updated_at AS query_updated_at,
			tb1.id AS chart_id,
			tb1.* 
		FROM
			%s AS tb1
			LEFT JOIN %s AS tb2 ON tb1.query_id = tb2.id 
		WHERE
			tb1.user_id = ? 
			LIMIT ? OFFSET ( ? - 1 ) * ?
		`

	rows, err := s.db.Raw(fmt.Sprintf(sql, tb1, tb2), userId, pageSize, page, pageSize).Rows()
	if err != nil {
		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	defer rows.Close()
	for rows.Next() {
		row := make(map[string]any)
		if err := s.db.ScanRows(rows, &row); err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		// convert chart string to structured chart
		chartConfig, ok := row["config"]
		if ok && chartConfig != nil {
			jsonChart := make(map[string]any, 0)
			if err := json.Unmarshal([]byte(chartConfig.(string)), &jsonChart); err != nil {
				log.Printf("unmarshal chart error: %v", err)
				continue
			}
			row["config"] = jsonChart
		}

		charts = append(charts, row)

	}

	sql = `
		SELECT COUNT
			( ID ) 
		FROM
			%s 
		WHERE
			user_id = ?
		`

	rows, err = s.db.Raw(fmt.Sprintf(sql, tb1), userId).Rows()
	if err != nil {
		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var totalCount int
	for rows.Next() {
		if err := s.db.ScanRows(rows, &totalCount); err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}
	}

	base.ResponseWithMap(ctx, map[string]any{
		"charts": charts,
		"total":  totalCount,
	})
}

func (s *Service) getCurrentUserQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		s.getUserQueryChart(ctx, userId)

	}
}

func (s *Service) getUserQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetUintParam(ctx, "userId")
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		s.getUserQueryChart(ctx, userId)

	}
}

func (s *Service) getUserQueryChart(ctx *gin.Context, userId uint) {
	id, err := base.GetUintParam(ctx, "id")
	if err != nil {
		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	needJoinQuery := true
	if _, err := base.GetUIntQuery(ctx, "query_id"); err != nil {
		if err == base.ErrQueryNotFound {
			needJoinQuery = false
		}
		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	tb1 := datamodel.ChartModel{}.TableName()
	tb2 := datamodel.QueryModel{}.TableName()

	var sql string
	if needJoinQuery {
		sql = `
			SELECT
				tb2.name AS query_name,
				tb2.description AS query_description,
				tb2.query,
				tb2.query_engine,
				tb2.is_privacy,
				tb2.unsaved,
				tb2.stars as query_stars,
				tb2.created_at AS query_created_at,
				tb2.updated_at AS query_updated_at,
				tb1.id AS chart_id,
				tb1.* 
			FROM
				%s AS tb1
				LEFT JOIN %s AS tb2 ON tb1.query_id = tb2.id 
			WHERE
				tb1.id = ? 
				AND tb1.user_id = ? 
			`
		sql = fmt.Sprintf(sql, tb1, tb2)
		fmt.Println(sql)

	} else {
		sql = `
			SELECT
				tb1.id AS chart_id,
				tb1.* 
			FROM
				%s AS tb1
			WHERE
				tb1.id = ? 
				AND tb1.user_id = ? 
			`
		sql = fmt.Sprintf(sql, tb1)
	}

	rows, err := s.db.Raw(sql, id, userId).Rows()
	if err != nil {
		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	defer rows.Close()

	var chart map[string]any
	for rows.Next() {
		if err := s.db.ScanRows(rows, &chart); err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		} else {
			break
		}
	}

	base.ResponseWithData(ctx, chart)
}

func (s *Service) createQueryHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var request datamodel.QueryModel
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, "bind error: %v", err)
			return
		}

		if len(request.QueryEngine) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "query engine is required")
			return
		}
		if len(request.Query) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "query is required")
			return
		}

		_, ok := s.engines[request.QueryEngine]
		if !ok {
			base.ResponseErr(ctx, http.StatusBadRequest, "The %s query engine unsupported now", request.QueryEngine)
			return
		}

		request.UserID = currentUserId

		if !request.Unsaved {
			if len(request.Name) == 0 {
				base.ResponseErr(ctx, http.StatusBadRequest, "name is required")
				return
			}
		} else {
			request.Name = "unsaved"
		}

		request.CreatedAt = time.Now()
		request.UpdatedAt = time.Now()

		result := s.db.Create(&request)
		if result.Error != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		// create or update statistics
		var statistics datamodel.UserStatistics
		result = s.db.Where("user_id", currentUserId).First(&statistics)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				statistics.Queries += 1
				statistics.UserId = currentUserId
				if err := s.db.Create(&statistics).Error; err != nil {
					base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
					return
				}
				ctx.JSON(http.StatusOK, Response{
					BaseResponse: base.BaseResponse{
						Success: true,
					},
					Data: request,
				})

				return
			}

			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		if err := s.db.Model(&statistics).Update("queries", statistics.Queries+1).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		ctx.JSON(http.StatusOK, Response{
			BaseResponse: base.BaseResponse{
				Success: true,
			},
			Data: request,
		})
	}
}

func (s *Service) updateHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		_, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		var request datamodel.QueryModel
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, "bind error: %v", err)
			return
		}

		if !s.checkUserQueryModelRequest(ctx, &request) {
			return
		}
		request.UpdatedAt = time.Now()
		request.Unsaved = false

		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := s.db.Save(&request).Error; err != nil {
				return err
			}

			if err := s.db.Save(&request.Charts).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, Response{
			BaseResponse: base.BaseResponse{
				Success: true,
			},
			Data: request,
		})
	}
}

func (s *Service) queryFavoriteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.queryFavorite(ctx, true)
	}
}

func (s *Service) queryUnfavoriteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.queryFavorite(ctx, false)
	}
}

func (s *Service) queryFavorite(ctx *gin.Context, star bool) {
	userId, err := base.GetCurrentUserId(ctx)
	if err != nil {
		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var request datamodel.UserQueryFavorites
	if err := ctx.ShouldBindJSON(&request); err != nil {
		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if request.QueryID == 0 {
		base.ResponseErr(ctx, http.StatusBadRequest, "query id is required")
		return
	}

	if request.QueryUserID == 0 {
		base.ResponseErr(ctx, http.StatusBadRequest, "query user id is required")
		return
	}

	var (
		find   datamodel.UserQueryFavorites
		finded bool = true
	)
	if err := s.db.Table(request.TableName()).Where("user_id = ? and query_id = ?", userId, request.QueryID).First(&find).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			finded = false
		} else {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if finded {
			find.UpdatedAt = time.Now()
			find.Stared = star
			if err = tx.Where("user_id = ? and query_id = ?", userId, find.QueryID).Save(&find).Error; err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return err
			}
		} else {
			find.UserID = userId
			find.QueryID = request.QueryID
			find.CreatedAt = time.Now()
			find.Stared = star
			if err := tx.Create(&find).Error; err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return err
			}
		}

		expr := "stars + ?"
		if !star {
			expr = "stars - ?"
		}
		if err := s.db.Model(&datamodel.UserStatistics{}).Where("user_id = ?", request.QueryUserID).Update("stars", gorm.Expr(expr, 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	base.ResponseWithData(ctx, find)
}

func (s *Service) Name() string {
	return Name
}

func (s *Service) RouteTables() []base.RouteTable {
	return []base.RouteTable{
		{
			Method:  "GET",
			Path:    s.group + "/engines",
			Handler: s.ListEnginesHandle(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/engines/:engineId/datasets/:datasetId",
			Handler: s.GetQueryEngineDatasetHandle(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/run",
			Handler: s.runHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/:id",
			Handler: s.getQueryHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group,
			Handler: s.listQueryHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/favorite",
			Handler: s.listFavoriteQueryHandler(),
		},
		{
			Method:  "POST",
			Path:    s.group,
			Handler: s.createQueryHandler(),
		},
		{
			Method:  "PUT",
			Path:    s.group,
			Handler: s.updateHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/charts",
			Handler: s.listCurrentUserQueryChartHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/charts/user/:userId",
			Handler: s.listUserQueryChartHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/chart/:id",
			Handler: s.getCurrentUserQueryChartHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/chart/:id/user/:userId",
			Handler: s.getUserQueryChartHandler(),
		},
		{
			Method:  "PUT",
			Path:    s.group + "/favorite",
			Handler: s.queryFavoriteHandler(),
		},
		{
			Method:  "PUT",
			Path:    s.group + "/unfavorite",
			Handler: s.queryUnfavoriteHandler(),
		},
	}
}
