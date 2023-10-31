package base

import "github.com/gin-gonic/gin"

type RouteTable struct {
	Method     string
	Path       string
	AllowGuest bool
	Regexp     string
	Handler    gin.HandlerFunc
}
