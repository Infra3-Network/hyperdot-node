package base

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	_                = iota
	Ok               // Response code for success
	Err              // Response code for error
	InvalidArguments // Response code for invalid arguments
)

// BaseResponse is a base response struct, it inclues
// success, error message and error code that can be
// used in all response.
type BaseResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"errorMessage"`
	ErrorCode    int    `json:"errorCode"`
}

// ResponseOk returns a BaseResponse with success true
func ResponseOk() BaseResponse {
	return BaseResponse{
		Success: true,
	}
}

// ResponseSuccess returns a BaseResponse with success true
func ResponseSuccess(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, BaseResponse{
		Success: true,
	})
}

// ResponseErr returns a BaseResponse with success false
func ResponseErr(ctx *gin.Context, code int, format string, args ...any) {
	ctx.JSON(code, BaseResponse{
		Success:      false,
		ErrorCode:    Err,
		ErrorMessage: fmt.Sprintf(format, args...),
	})
}

// ResponseWithMap returns a BaseResponse with success true and map data
func ResponseWithMap(ctx *gin.Context, data map[string]any) {
	ctx.JSON(200, gin.H{
		"success": true,
		"data":    data,
	})
}

// ResponseWithData returns a BaseResponse with success true and data
func ResponseWithData(ctx *gin.Context, data any) {
	ctx.JSON(200, gin.H{
		"success": true,
		"data":    data,
	})
}
