package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"txing-ai/internal/observability"
)

// Metrics HTTP RED 指标中间件：请求计数 + 时延。
// 仅统计 /api 前缀业务流量（静态资源/Swagger//metrics 自身不计）；
// route 取 gin 路由模板（c.FullPath）防标签基数爆炸，未匹配记 "unnamed"
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !observability.Enabled() || !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unnamed"
		}
		observability.HTTPRequest(route, c.Request.Method, c.Writer.Status())
		observability.HTTPDuration(route, time.Since(start).Seconds())
	}
}
