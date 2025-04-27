package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"txing-ai/internal/enum"
	"txing-ai/internal/global"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/utils"
)

const TokenKey = "Authorization"

// 认证鉴权中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Token 格式： Authorization: Bearer <token>
		// 是否带了 Token
		token := ctx.Request.Header.Get(TokenKey)
		if token == "" {
			// 尝试从 query 中获取 token
			token = ctx.Query(TokenKey)
			if token == "" {
				log.Error("Token is missing")
				// 如果是路径包含 /chat/ws，直接放行
				if strings.Contains(ctx.Request.URL.Path, "/chat/ws") {
					log.Info("Websocket connection, allow have no token")
					ctx.Next()
					return
				}
				utils.ErrorWithHttpCode(ctx, http.StatusUnauthorized, global.CodeNotLogin, nil)
				ctx.Abort()
				return
			}
		}
		// Token 格式是否正确
		parts := strings.SplitN(token, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Error("Token format is invalid")
			utils.ErrorWithHttpCode(ctx, http.StatusUnauthorized, global.CodeNotLogin, nil)
			ctx.Abort()
			return
		}
		// 验证 token
		claims, err := utils.VerifyToken(parts[1])
		if err != nil {
			log.Error("Token is invalid")
			utils.ErrorWithHttpCode(ctx, http.StatusUnauthorized, global.CodeNotLogin, nil)
			ctx.Abort()
			return
		}
		// 和 Redis 中存储的进行对比，是否一致，如果不一致，则表示已在其他地方登陆，禁止访问
		// 暂时不做单一用户多端登陆的限制
		//result, err := redisClient.Get(context.Background(),
		//	common.GetUserTokenKey(claims.UserID, ctx.RemoteIP())).
		//	Result()
		//if err != nil || result != parts[1] {
		//	if err != nil {
		//		zap.L().Error("Token is invalid", zap.Error(err))
		//	} else {
		//		zap.L().Error("User has logged in elsewhere",
		//			zap.Int64("user_id", claims.UserID), zap.String("ip", ctx.RemoteIP()))
		//	}
		//	common.ErrorWithCode(ctx, common.CodeNotLogin)
		//	ctx.Abort()
		//	return
		//}

		// 如果请求路径是 /api/admin 开头，则需要验证是否是管理员
		if strings.HasPrefix(ctx.Request.URL.Path, "/api/admin") {
			if claims.Role != enum.UserTypeSuper {
				log.Error("User is not admin, cannot access admin api")
				utils.ErrorWithHttpCode(ctx, http.StatusForbidden, global.CodeNotPermission, nil)
				ctx.Abort()
				return
			}
		}

		// 验证通过，设置当前用户 ID 到上下文中
		ctx.Set("userId", claims.UserID)
		ctx.Set("role", claims.Role)
		// 放行
		ctx.Next()
	}
}
