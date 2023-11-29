package base

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JwtAuthMiddleware is a middleware to verify jwt token
func JwtAuthMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/apis/v1/user/auth/login" ||
			ctx.Request.URL.Path == "/apis/v1/user/auth/createAccount" ||
			ctx.Request.URL.Path == "/apis/v1/file" ||
			strings.Contains(ctx.Request.URL.Path, "swager") ||
			strings.Contains(ctx.Request.URL.Path, "/apis/v1/file") {
			ctx.Next()
			return
		}

		authHeader := ctx.Request.Header.Get("authorization")
		if authHeader == "" {
			ResponseErr(ctx, http.StatusUnauthorized, "invalid authorization")
			ctx.Abort()
			return
		}

		claims, err := VerifyJwtToken(authHeader)
		if err != nil {
			ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Next()
	}
}
