package apis

import (
	"fmt"

	"infra-3.xyz/hyperdot-node/internal/apis/service/dashboard"
	"infra-3.xyz/hyperdot-node/internal/apis/service/file"
	"infra-3.xyz/hyperdot-node/internal/apis/service/query"
	"infra-3.xyz/hyperdot-node/internal/apis/service/system"
	"infra-3.xyz/hyperdot-node/internal/clients"
	"infra-3.xyz/hyperdot-node/internal/dataengine"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	docs "infra-3.xyz/hyperdot-node/docs"

	"infra-3.xyz/hyperdot-node/internal/apis/base"
	"infra-3.xyz/hyperdot-node/internal/apis/service/user"
	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/store"
)

const (
	BASE            = "/apis"
	V1              = "v1"
	CURRENT_VERSION = V1
)

// Router is a router interface can be collect routers of service and
// register to gin engine
type Router interface {
	Name() string
	RouteTables() []base.RouteTable
}

// RouterBuilder is a builder to build gin engine
type RouterBuilder struct {
	enableQuery bool
	boltStore   *store.BoltStore
	db          *gorm.DB
	cfg         *common.Config
	s3Client    *clients.SimpleS3Cliet
	engines     map[string]dataengine.QueryEngine
}

// NewRouterBuilder creates a new RouterBuilder
func NewRouterBuilder(
	boltStore *store.BoltStore,
	cfg *common.Config,
	db *gorm.DB,
	engines map[string]dataengine.QueryEngine,
	s3Client *clients.SimpleS3Cliet,
) *RouterBuilder {
	return &RouterBuilder{
		enableQuery: true,
		boltStore:   boltStore,
		db:          db,
		cfg:         cfg,
		engines:     engines,
		s3Client:    s3Client,
	}
}

// Build builds gin engine
func (r *RouterBuilder) Build() (*gin.Engine, error) {
	engine := gin.Default()
	versionUrl := fmt.Sprintf("%s/%s", BASE, CURRENT_VERSION)
	docs.SwaggerInfo.BasePath = "/apis/v1"
	engine.Use(base.JwtAuthMiddleware())
	router := engine.Group(versionUrl)
	{
		var svcs []Router
		svcs = append(svcs, system.New(r.cfg))
		svcs = append(svcs, query.New(r.boltStore, r.cfg, r.db, r.engines))
		svcs = append(svcs, dashboard.New(r.db))
		svcs = append(svcs, user.New(r.db, r.engines, r.s3Client))
		svcs = append(svcs, file.New(r.s3Client))
		for _, svc := range svcs {
			for _, table := range svc.RouteTables() {
				router.Handle(table.Method, table.Path, table.Handler)
			}
		}
	}
	router.GET("/swager/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return engine, nil
}
