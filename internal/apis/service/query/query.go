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

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/cache"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/dataengine"
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
		// ctx.JSON(200, base.GetQueryEngineDatasetResponse{
		// 	BaseResponse: base.ResponseOk(),
		// 	Data: struct {
		// 		Id          string                                   `json:"id"`
		// 		Chains      map[int]datamodel.Chain                  `json:"chains"`
		// 		RelayChains map[string]*datamodel.RelayChainMetadata `json:"relayChains"`
		// 		ChainTables map[int][]string                         `json:"chainTables"`
		// 	}(struct {
		// 		Id          string
		// 		Chains      map[int]datamodel.Chain
		// 		RelayChains map[string]*datamodel.RelayChainMetadata
		// 		ChainTables map[int][]string
		// 	}{Id: data.Id, Chains: data.Chains, RelayChains: data.RelayChains, ChainTables: chainTables}),
		// })
	}
}

// @Summary run query
// @Description run query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param body body RequestRunQuery true "query body"
// @Success 200 {object} ResponseRun
// @Router /query/run [get]
func (s *Service) RunHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var request RequestRunQuery
		if err := ctx.ShouldBindJSON(&request); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, "bind error: %v", err)
			return
		}

		_, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if len(request.Query) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "query engine is required")
			return
		}
		if len(request.Engine) == 0 {
			base.ResponseErr(ctx, http.StatusBadRequest, "query is required")
			return
		}

		engine, ok := s.engines[request.Engine]
		if !ok {
			base.ResponseErr(ctx, http.StatusBadRequest, "The %s query engine unsupported now", request.Engine)
			return
		}

		iter, err := engine.Run(context.Background(), request.Query)
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

		ctx.JSON(http.StatusOK, ResponseRun{
			BaseResponse: base.BaseResponse{
				Success: true,
			},
			Data: ResponseRunData{
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

// @Summary get query
// @Description get query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param id path int true "query id"
// @Success 200 {object} Response
// @Router /query/:id [get]
func (s *Service) GetQueryHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		var query datamodel.QueryModel
		result := s.db.First(&query, id)
		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				base.ResponseErr(ctx, http.StatusNotFound, "query not found")
				return
			}
			base.ResponseErr(ctx, http.StatusInternalServerError, result.Error.Error())
			return
		}

		if err := s.db.Where("query_id = ? AND user_id = ?", query.ID, query.UserID).Order("id ASC").Find(&query.Charts).Error; err != nil {
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

// @Summary list query
// @Description list query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param user_id query int false "user_id"
// @Param time_range query string false "time_range"
// @Param order query string false "order"
// @Success 200
// @Router /query [get]
func (s *Service) ListQueryHandler() gin.HandlerFunc {
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
	}
}

// @Summary list favorite query
// @Description list favorite query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param user_id query int false "user_id"
// @Param time_range query string false "time_range"
// @Param order query string false "order"
// @Success 200
// @Router /query/favorite [get]
func (s *Service) ListFavoriteQueryHandler() gin.HandlerFunc {
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

// @Summary list browse query
// @Description list browse query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param user_id query int false "user_id"
// @Param time_range query string false "time_range"
// @Param order query string false "order"
// @Success 200
// @Router /query/browse [get]
func (s *Service) ListBrowseQueryHandler() gin.HandlerFunc {
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

		queryRaw, countRaw, err := s.prepateBrowseUserListSQL(params)
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

// @Summary list current logined user charts
// @Description list current logined user charts
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param time_range query string false "time_range"
// @Param order query string false "order"
// @Success 200
// @Router /query/charts [get]
func (s *Service) ListCurrentUserQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		s.listUserQueryChart(ctx, userId)
	}
}

// @Summary list user charts
// @Description list user charts
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param userId path int true "user id"
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param time_range query string false "time_range"
// @Param order query string false "order"
// @Success 200
// @Router /query/charts/user/:userId [get]
func (s *Service) ListUserQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userId, err := base.GetUintParam(ctx, "userId")
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		s.listUserQueryChart(ctx, userId)
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
		} else {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}
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

// @Summary get current logined user query chart
// @Description get current logined user query chart
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param id path int true "chart id"
// @Success 200
// @Router /query/chart/:id [get]
func (s *Service) GetCurrentUserQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		s.getUserQueryChart(ctx, userId)

	}
}

// @Summary get user query chart
// @Description get user query chart
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param id path int true "chart id"
// @Param userId path int true "user id"
// @Success 200
// @Router /query/chart/:id/user/:userId [get]
func (s *Service) GetUserQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetUintParam(ctx, "userId")
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		s.getUserQueryChart(ctx, userId)

	}
}

// @Summary delete query chart
// @Description delete query chart
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param id path int true "chart id"
// @Success 200
// @Router /query/chart/:id [delete]
func (s *Service) DeleteQueryChartHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := base.GetUintParam(ctx, "id")
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.db.Where("id = ?", id).Delete(&datamodel.ChartModel{}).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		base.ResponseSuccess(ctx)
	}
}

// @Summary create query chart
// @Description create query chart
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param body body datamodel.QueryModel true "body"
// @Success 200 {object} Response
// @Router /query/chart [post]
func (s *Service) CreateQueryHandler() gin.HandlerFunc {
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

// @Summary update query
// @Description update query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param body body datamodel.QueryModel true "body"
// @Success 200 {object} Response
// @Router /query [put]
func (s *Service) UpdateQueryHandler() gin.HandlerFunc {
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

// @Summary delete query
// @Description delete query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param id path int true "query id"
// @Success 200
// @Router /query/:id [delete]
func (s *Service) DeleteQueryHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		id, err := base.GetUintParam(ctx, "id")
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		// first delete related charts and then delete query using transaction
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := s.db.Where("query_id = ?", id).Delete(&datamodel.ChartModel{}).Error; err != nil {
				return err
			}

			if err := s.db.Where("id = ? and user_id = ?", id, userId).Delete(&datamodel.QueryModel{}).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		base.ResponseSuccess(ctx)
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

// @Summary user favorite query
// @Description user favorite query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param body body datamodel.UserQueryFavorites true "body"
// @Success 200
// @Router /query/favorite [put]
func (s *Service) QueryFavoriteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.queryFavorite(ctx, true)
	}
}

// @Summary user unfavorite query
// @Description user unfavorite query
// @Tags query apis
// @Accept application/json
// @Produce application/json
// @Param body body datamodel.UserQueryFavorites true "body"
// @Success 200
// @Router /query/unfavorite [put]
func (s *Service) QueryUnfavoriteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.queryFavorite(ctx, false)
	}
}

func (s *Service) Name() string {
	return Name
}

func (s *Service) RouteTables() []base.RouteTable {
	return []base.RouteTable{
		{
			Method:  "POST",
			Path:    s.group + "/run",
			Handler: s.RunHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/:id",
			Handler: s.GetQueryHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group,
			Handler: s.ListQueryHandler(),
		},
		{
			Method:  "POST",
			Path:    s.group,
			Handler: s.CreateQueryHandler(),
		},
		{
			Method:  "PUT",
			Path:    s.group,
			Handler: s.UpdateQueryHandler(),
		},
		{
			Method:  "DELETE",
			Path:    s.group + "/:id",
			Handler: s.DeleteQueryHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/favorite",
			Handler: s.ListFavoriteQueryHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/browse",
			Handler: s.ListBrowseQueryHandler(),
		},

		{
			Method:  "GET",
			Path:    s.group + "/charts",
			Handler: s.ListCurrentUserQueryChartHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/charts/user/:userId",
			Handler: s.ListUserQueryChartHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/chart/:id",
			Handler: s.GetCurrentUserQueryChartHandler(),
		},
		{
			Method:  "GET",
			Path:    s.group + "/chart/:id/user/:userId",
			Handler: s.GetUserQueryChartHandler(),
		},
		{
			Method:  "DELETE",
			Path:    s.group + "/chart/:id/",
			Handler: s.DeleteQueryChartHandler(),
		},
		{
			Method:  "PUT",
			Path:    s.group + "/favorite",
			Handler: s.QueryFavoriteHandler(),
		},
		{
			Method:  "PUT",
			Path:    s.group + "/unfavorite",
			Handler: s.QueryUnfavoriteHandler(),
		},
	}
}
