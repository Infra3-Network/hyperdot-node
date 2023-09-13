package base

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// JwtAuthMiddleware 基于JWT的认证中间件--验证用户是否登录
func JwtAuthMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// if url is /apis/v1/user/auth/login, next
		if ctx.Request.URL.Path == "/apis/v1/user/auth/login" {
			ctx.Next()
			return
		}

		authHeader := ctx.Request.Header.Get("authorization")
		if authHeader == "" {
			ResponseErr(ctx, http.StatusUnauthorized, "invalid authorization")
			ctx.Abort()
			return
		}

		//parts := strings.Split(authHeader, ".")
		//if len(parts) != 3 {
		//	ResponseErr(ctx, http.StatusUnauthorized, "invalid authorization format")
		//	ctx.Abort()
		//	return
		//}

		claims, err := VerifyJwtToken(authHeader)
		if err != nil {
			ResponseErr(ctx, http.StatusUnauthorized, err.Error())
			ctx.Abort()
			return
		}

		ctx.Set("username", claims.UserID)
		ctx.Next()
	}
}
