package base

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

var ErrQueryNotFound = errors.New("query not found")

func GetCurrentUserId(ctx *gin.Context) (uint, error) {
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

func GetUintParam(ctx *gin.Context, key string) (uint, error) {
	v := ctx.Param(key)
	if len(v) == 0 {
		return 0, fmt.Errorf("%s is required", key)
	}

	res, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}

	return uint(res), nil

}

func GetIntQuery(ctx *gin.Context, key string) (int, error) {
	v := ctx.Query(key)
	if len(v) == 0 {
		return 0, fmt.Errorf("%s is required", key)
	}

	res, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return res, nil

}

func GetUIntQuery(ctx *gin.Context, key string) (uint, error) {
	v := ctx.Query(key)
	if len(v) == 0 {
		return 0, ErrQueryNotFound
	}

	res, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return uint(res), nil

}

func GetUIntQueryRequired(ctx *gin.Context, key string) (uint, error) {
	v := ctx.Query(key)
	if len(v) == 0 {
		return 0, fmt.Errorf("%s is required", key)
	}

	res, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return uint(res), nil

}
