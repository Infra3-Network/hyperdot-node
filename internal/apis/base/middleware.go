package base

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JwtAuthMiddleware 基于JWT的认证中间件--验证用户是否登录
func JwtAuthMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		// if url is /apis/v1/user/auth/login, next
		if ctx.Request.URL.Path == "/apis/v1/user/auth/login" || ctx.Request.URL.Path == "/apis/v1/user/auth/createAccount" {
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

		ctx.Set("user_id", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Next()
	}
}
