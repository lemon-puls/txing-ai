package route

import (
	"github.com/gin-gonic/gin"
	"txing-ai/internal/controller/agent"
	"txing-ai/internal/controller/captcha"
	"txing-ai/internal/controller/channel"
	"txing-ai/internal/controller/chat"
	"txing-ai/internal/controller/cos"
	"txing-ai/internal/controller/model"
	"txing-ai/internal/controller/preset"
	"txing-ai/internal/controller/user"
	"txing-ai/internal/controller/website"
	"txing-ai/internal/iface"
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

	chat.Register(group.Group("/chat"))

	channel.Register(group)

	model.Register(group)

	preset.Register(group)

	website.Register(group)

	// 验证码相关路由
	captcha.Register(group.Group("/captcha"))

	// agent 相关路由
	agent.Register(group.Group("/agent"))

	// COS 相关路由
	cos.Register(group.Group("/cos"))

	// 注册Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
