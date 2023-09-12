package apis

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	docs "infra-3.xyz/hyperdot-node/docs"
	"net/http"

	"infra-3.xyz/hyperdot-node/internal/common"
	"infra-3.xyz/hyperdot-node/internal/store"
)

const (
	BASE            = "/apis"
	V1              = "v1"
	CURRENT_VERSION = V1
)

type RouteTable struct {
	Method  string
	Group   string
	Path    string
	Handler gin.HandlerFunc
}

type Router interface {
	Name() string
	RouteTables() []RouteTable
}

type afterMiddlewareWriter struct {
	gin.ResponseWriter
}

func CorsResponse(r *http.Request, w gin.ResponseWriter) {
	w.Header().Add("X-TIME", "time")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Methods", "POST,GET,OPTIONS,DELETE,PUT,HEAD,PATCH")
	w.Header().Add("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Add("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
}

func AfterMiddleware(c *gin.Context) {
	c.Writer = &afterMiddlewareWriter{c.Writer}
	c.Next()
}

type RouterBuilder struct {
	enableQuery bool
	boltStore   *store.BoltStore
	cfg         *common.Config
}

func NewRouterBuilder(boltStore *store.BoltStore, cfg *common.Config) *RouterBuilder {
	return &RouterBuilder{
		enableQuery: true,
		boltStore:   boltStore,
		cfg:         cfg,
	}
}

func (r *RouterBuilder) Build() (*gin.Engine, error) {
	engine := gin.Default()
	versionUrl := fmt.Sprintf("%s/%s", BASE, CURRENT_VERSION)
	docs.SwaggerInfo.BasePath = "/apis/v1"
	engine.Use(cors.Default())
	router := engine.Group(versionUrl)
	{
		var svcs []Router
		if r.enableQuery {
			svc, err := NewQueryService(r.boltStore, r.cfg)
			if err != nil {
				return nil, err
			}
			svcs = append(svcs, svc)
		}

		for _, svc := range svcs {
			for _, table := range svc.RouteTables() {
				router.Handle(table.Method, table.Path, table.Handler)
			}
		}
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	//

	return engine, nil
}
