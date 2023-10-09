package dashboard

import (
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
