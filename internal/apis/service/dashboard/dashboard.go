package dashboard

import (
	"fmt"
	"net/http"
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

func (s *Service) listDashboardHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := base.GetCurrentUserId(ctx)
		if err != nil {
			base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
			return
		}

		var (
			page     uint
			pageSize uint
			userId   uint
		)

		page, err = base.GetUIntQuery(ctx, "page")
		if err != nil {
			if err == base.ErrQueryNotFound {
				page = 1
			} else {
				base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
				return
			}
		}

		pageSize, err = base.GetUIntQuery(ctx, "page_size")
		if err != nil {
			if err == base.ErrQueryNotFound {
				pageSize = 10
			} else {
				base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
				return
			}
		}

		userId, err = base.GetUIntQuery(ctx, "user_id")
		if err != nil {
			if err == base.ErrQueryNotFound {
				userId = 0
			} else {
				base.ResponseErr(ctx, http.StatusBadRequest, err.Error())
				return
			}
		}

		tb1 := datamodel.DashboardModel{}.TableName()
		tb2 := datamodel.UserModel{}.TableName()

		var raw *gorm.DB
		if userId == 0 {
			fmt.Println("userId == 0")
			sql := `
			SELECT
				tb1.*,
				tb2.username,
				tb2.username,
				tb2.email,
				tb2.icon_url
			FROM
				%s AS tb1
				LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
			WHERE
				tb1.is_privacy = FALSE 
			ORDER BY
				updated_at DESC 
				LIMIT ? OFFSET ( ? - 1 ) * ?
		`
			sql = fmt.Sprintf(sql, tb1, tb2)
			raw = s.db.Raw(sql, pageSize, page, pageSize)
		} else {
			sql := `
			SELECT
				tb1.*,
				tb2.username,
				tb2.username,
				tb2.email,
				tb2.icon_url
			FROM
				%s AS tb1
				LEFT JOIN %s AS tb2 ON tb1.user_id = tb2.id
			WHERE
				tb1.is_privacy = FALSE 
				AND tb1.user_id = ?
			ORDER BY
				updated_at DESC 
				LIMIT ? OFFSET ( ? - 1 ) * ?
		`
			sql = fmt.Sprintf(sql, tb1, tb2)
			raw = s.db.Raw(sql, userId, pageSize, page, pageSize)
		}

		rows, err := raw.Rows()
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
		if userId == 0 {
			sql := `
			SELECT COUNT(tb1.id)
			FROM
				%s AS tb1
			WHERE
				tb1.is_privacy = FALSE;
			`
			sql = fmt.Sprintf(sql, tb1)
			raw = s.db.Raw(sql)
		} else {
			sql := `
			SELECT COUNT(tb1.id)
			FROM
				%s AS tb1
			WHERE
				tb1.is_privacy = FALSE 
				AND tb1.user_id = ?
		`
			sql = fmt.Sprintf(sql, tb1)
			raw = s.db.Raw(sql, userId)
		}

		if rows, err = raw.Rows(); err != nil {
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
