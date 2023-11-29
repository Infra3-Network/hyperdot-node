package base

import "github.com/gin-gonic/gin"

// RouterTable is a table to store route information
type RouteTable struct {
	Method     string
	Path       string
	AllowGuest bool
	Regexp     string
	Handler    gin.HandlerFunc
}
