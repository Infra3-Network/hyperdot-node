package dashboard

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/datamodel"
)

const Name = "Dashboard"

type Service struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

func (s *Service) Name() string {
	return Name
}

func (s *Service) RouteTables() []base.RouteTable {
	group := "dashboard"
	return []base.RouteTable{
		{
			Method:  "GET",
			Path:    group + "/:id",
			Handler: s.getDashboardHandler(),
		},
		{
			Method:  "GET",
			Path:    group,
			Handler: s.listDashboardHandler(),
		},
		{
			Method:  "POST",
			Path:    group,
			Handler: s.createDashboardHandler(),
		},
		{
			Method:  "PUT",
			Path:    group,
			Handler: s.updateDashboardHandler(),
		},
		{
			Method:  "DELETE",
			Path:    group + "/:id",
			Handler: s.deleteDashboardHandler(),
		},
		{
			Method:  "GET",
			Path:    group + "/favorite",
			Handler: s.listFavoriteDashboardHandler(),
		},
		{
			Method:  "GET",
			Path:    group + "/browse",
			Handler: s.listBrowseUserDashboardHandler(),
		},
		{
			Method:  "GET",
			Path:    group + "/tag/populars",
			Handler: s.listPopularDashboardTags(),
		},

		{
			Method:  "PUT",
			Path:    group + "/favorite",
			Handler: s.dashboardFavoriteHandler(),
		},
		{
			Method:  "PUT",
			Path:    group + "/unfavorite",
			Handler: s.dashboardUnfavoriteHandler(),
		},
		{
			Method:  "DELETE",
			Path:    group + "/panel/:panelId",
			Handler: s.removeDashboardPanelHandler(),
		},
	}
}

func (s *Service) getDashboardHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		id, err := base.GetUintParam(ctx, "id")
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var dashboard datamodel.DashboardModel
		dashboard.ID = id
		if err := s.db.First(&dashboard).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				base.ResponseErr(ctx, http.StatusNotFound, err.Error())
				return
			}

			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		if err := s.db.Where("dashboard_id = ?", dashboard.ID).Find(&dashboard.Panels).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		base.ResponseWithData(ctx, dashboard)
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

func (s *Service) listDashboardHandler() gin.HandlerFunc {
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

		var dashboards []map[string]interface{}
		for rows.Next() {
			data := make(map[string]interface{})
			if err := s.db.ScanRows(rows, &data); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			dashboards = append(dashboards, data)
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
			"dashboards": dashboards,
			"total":      total,
		})

	}
}

func (s *Service) listFavoriteDashboardHandler() gin.HandlerFunc {
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

		var dashboards []map[string]interface{}
		for rows.Next() {
			data := make(map[string]interface{})
			if err := s.db.ScanRows(rows, &data); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			dashboards = append(dashboards, data)
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
			"dashboards": dashboards,
			"total":      total,
		})
	}
}

func (s *Service) listPopularDashboardTags() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limit, err := base.GetUIntQuery(ctx, "limit")
		if err != nil {
			if err == base.ErrQueryNotFound {
				limit = 10
			}
		}

		raw, err := s.preparePopularDashboardTagsSQL(limit)
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		rows, err := raw.Rows()
		if err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		defer rows.Close()

		var tagRows []map[string]interface{}
		for rows.Next() {
			data := make(map[string]interface{})
			if err := s.db.ScanRows(rows, &data); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			tagRows = append(tagRows, data)
		}

		tag2count := make(map[string]int64, 0)
		for _, tagRow := range tagRows {
			v, ok := tagRow["tags"]
			if !ok || v == nil {
				continue
			}

			vstr, ok := v.(string)
			if !ok || len(vstr) == 0 {
				continue
			}

			tags := strings.Split(vstr, ",")
			if len(tags) == 0 {
				continue
			}

			// init map by tags
			for _, tag := range tags {
				if _, ok := tag2count[tag]; !ok {
					tag2count[tag] = 0
				}
			}

			v, ok = tagRow["favorites_count"]
			if !ok || v == nil {
				continue
			}

			vint, ok := v.(int64)
			if !ok || vint == 0 {
				continue
			}

			for _, tag := range tags {
				tag2count[tag] += vint
			}
		}

		base.ResponseWithData(ctx, tag2count)
	}

}

func (s *Service) listBrowseUserDashboardHandler() gin.HandlerFunc {
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

		var dashboards []map[string]interface{}
		for rows.Next() {
			data := make(map[string]interface{})
			if err := s.db.ScanRows(rows, &data); err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}

			dashboards = append(dashboards, data)
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
			"dashboards": dashboards,
			"total":      total,
		})
	}
}

func (s *Service) createDashboardHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		var req datamodel.DashboardModel
		if err := ctx.ShouldBindJSON(&req); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}
		req.UserID = userId
		req.CreatedAt = time.Now()

		if err := s.db.Create(&req).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		base.ResponseWithData(ctx, req)
	}
}

func (s *Service) updateDashboardHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			return
		}

		var req datamodel.DashboardModel
		if err := ctx.ShouldBindJSON(&req); err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if req.UserID != userId {
			base.ResponseErr(ctx, http.StatusUnauthorized, "You are not the owner of this dashboard")
			return
		}

		req.UpdatedAt = time.Now()

		if err := s.db.Save(&req).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		for _, panel := range req.Panels {
			panel.DashboardID = req.ID
			if err := s.db.Save(&panel).Error; err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return
			}
		}

		base.ResponseWithData(ctx, req)
	}
}

func (s *Service) deleteDashboardHandler() gin.HandlerFunc {
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

		// delete panels and then delete dashboard using transaction
		err = s.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("dashboard_id = ?", id).Delete(&datamodel.DashboardPanelModel{}).Error; err != nil {
				return err
			}

			if err := tx.Where("id = ? and user_id = ?", id, userId).Delete(&datamodel.DashboardModel{}).Error; err != nil {
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

func (s *Service) dashboardFavoriteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.dashboardFavorite(ctx, true)
	}
}

func (s *Service) dashboardUnfavoriteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.dashboardFavorite(ctx, false)
	}
}

func (s *Service) dashboardFavorite(ctx *gin.Context, star bool) {
	userId, err := base.GetCurrentUserId(ctx)
	if err != nil {
		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var request datamodel.UserDashboardFavorites
	if err := ctx.ShouldBindJSON(&request); err != nil {
		base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if request.DashboardID == 0 {
		base.ResponseErr(ctx, http.StatusBadRequest, "dashboard id is required")
		return
	}

	if request.DashboardUserID == 0 {
		base.ResponseErr(ctx, http.StatusBadRequest, "dashboard user id is required")
		return
	}

	var (
		find   datamodel.UserDashboardFavorites
		finded bool = true
	)
	if err := s.db.Table(request.TableName()).Where("user_id = ? and dashboard_id = ?", userId, request.DashboardID).First(&find).Error; err != nil {
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
			if err = tx.Where("user_id = ? and dashboard_id = ?", userId, find.DashboardID).Save(&find).Error; err != nil {
				base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
				return err
			}
		} else {
			find.UserID = userId
			find.DashboardID = request.DashboardID
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
		if err := s.db.Model(&datamodel.UserStatistics{}).Where("user_id = ?", request.DashboardUserID).Update("stars", gorm.Expr(expr, 1)).Error; err != nil {
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

func (s *Service) removeDashboardPanelHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		panelId, err := base.GetUintParam(ctx, "panelId")
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.db.Where("id = ?", panelId).Delete(&datamodel.DashboardPanelModel{}).Error; err != nil {
			base.ResponseErr(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		base.ResponseSuccess(ctx)
	}
}
