package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
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

type RouterBuilder struct {
	enableQuery bool
}

func NewRouterBuilder() *RouterBuilder {
	return &RouterBuilder{
		enableQuery: true,
	}
}

func (r *RouterBuilder) Build() *gin.Engine {
	engine := gin.Default()
	versionUrl := fmt.Sprintf("%s/%s", BASE, CURRENT_VERSION)
	router := engine.Group(versionUrl)
	{
		var svcs []Router
		if r.enableQuery {
			svcs = append(svcs, NewQueryService())
		}

		for _, svc := range svcs {
			for _, table := range svc.RouteTables() {
				router.Handle(table.Method, table.Path, table.Handler)
			}
		}
	}

	return engine
}
