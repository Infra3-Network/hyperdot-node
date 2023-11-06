package base

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ErrQueryNotFound is returned when the query is not found.
var ErrQueryNotFound = errors.New("query not found")

// GetCurrentUser get current login user id
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

// GetUintParam get uint param from gin context
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

// GetIntQuery get int query from gin context
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

// GetUIntQuery get uint query from gin context.
// If the query is not found, return special error and the caller should handle it.
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

// GetUIntQueryRequired get uint query from gin context.
// If the query is not found, return error.
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

// GetStringQuery get string query from gin context.
// If the query is not found, return special error and the caller should handle it.
func GetStringQuery(ctx *gin.Context, key string) (string, error) {
	v := ctx.Query(key)
	if len(v) == 0 {
		return "", ErrQueryNotFound
	}

	return v, nil
}
