package utils

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// 添加获取Bearer Token的方法
func GetBearerToken(ctx *gin.Context) string {
	authHeader := ctx.GetHeader("Authorization")
	if len(authHeader) > 7 && strings.ToUpper(authHeader[0:6]) == "BEARER" {
		return authHeader[7:]
	}
	return ""
}
