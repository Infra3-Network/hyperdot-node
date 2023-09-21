package base

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func CurrentUserId(ctx *gin.Context) (uint, error) {
	v, ok := ctx.Get("user_id")
	if !ok {
		return 0, fmt.Errorf("user not login")
	}
	currentLoginUserId, ok := v.(uint)
	if !ok {
		return 0, fmt.Errorf("user not login")
	}

	return currentLoginUserId, nil
}
