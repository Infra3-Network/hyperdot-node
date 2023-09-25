package base

import "github.com/gin-gonic/gin"

type RouteTable struct {
	Method  string
	Path    string
	Handler gin.HandlerFunc
}
