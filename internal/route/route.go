package route

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"txing-ai/internal/controller/aboutme"
	"txing-ai/internal/controller/agent"
	"txing-ai/internal/controller/captcha"
	"txing-ai/internal/controller/channel"
	"txing-ai/internal/controller/chat"
	"txing-ai/internal/controller/cos"
	"txing-ai/internal/controller/dashboard"
	"txing-ai/internal/controller/file"
	"txing-ai/internal/controller/model"
	"txing-ai/internal/controller/ops"
	"txing-ai/internal/controller/preset"
	"txing-ai/internal/controller/user"
	"txing-ai/internal/controller/website"
	"txing-ai/internal/controller/wiki"
	"txing-ai/internal/controller/workflow"
	"txing-ai/internal/iface"
	"txing-ai/internal/observability"
	"txing-ai/static"

	_ "txing-ai/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Register(router gin.IRouter, res iface.ResourceProvider) {

	// 注册前端页面
	static.Register(router.Group("/dash"))

	// 全局的中间件，如 Swagger 的注册
	group := router.Group("/api")

	// 用户相关路由
	user.Register(group)

	chat.Register(group.Group("/chat"), res)

	channel.Register(group)

	model.Register(group)

	preset.Register(group)

	website.Register(group)

	// 运营助手相关路由（管理后台 AI 助手）
	ops.Register(group)

	workflow.Register(group, res)

	// 验证码相关路由
	captcha.Register(group.Group("/captcha"))

	// agent 相关路由
	agent.Register(group.Group("/agent"))

	// COS 相关路由
	cos.Register(group.Group("/cos"))

	// 文件上传下载相关路由
	file.Register(group)

	// 关于我页面路由（公开 + 管理后台）
	aboutme.Register(group)

	// LLM Wiki 知识库（公开问答 + 管理后台）
	wiki.Register(group)

	// 管理后台控制台（可观测性 + 运营统计）
	dashboard.Register(group)

	// 可观测性外部抓取端点：标准 Prometheus exposition（observability.external.enabled 时注册；
	// 可选 Bearer token 校验，不走 JWT，便于 Prometheus/VictoriaMetrics 直接抓取）
	if h := observability.PromHTTPHandler(); h != nil {
		router.GET("/metrics", func(c *gin.Context) {
			if !observability.CheckMetricsToken(c.GetHeader("Authorization")) {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			h.ServeHTTP(c.Writer, c.Request)
		})
	}

	// 注册Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
