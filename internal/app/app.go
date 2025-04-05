package app

import (
	"context"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"txing-ai/internal/global"
	"txing-ai/internal/global/config"
	"txing-ai/internal/global/logging/log"
	"txing-ai/internal/iface"
	"txing-ai/internal/middleware"
	"txing-ai/internal/route"

	"github.com/gin-gonic/gin"
)

type Server interface {
	Start()
	Shutdown(context.Context) error
}

type server struct {
	impl        *http.Server
	resProvider iface.ResourceProvider
}

var _ Server = (*server)(nil)

func (s *server) Start() {

	go func() {
		// TODO 加 recover 处理
		if err := s.impl.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("listen and serve error", zap.Error(err))
		}
		log.Info("server closed")
	}()
}

func (s *server) Shutdown(ctx context.Context) error {
	err := s.resProvider.GetRedisClient().Close()
	if err != nil {
		log.Error("redis client close error", zap.Error(err))
	}

	return s.impl.Shutdown(ctx)
}

func New(ctx context.Context, appConfig *global.AppConfig) Server {
	// 加载必要的资源，包括配置、日志、数据库连接等
	engine := gin.Default() // Default() 带了部分默认配置，要完全自定义使用 gin.New()

	// 初始化 DB
	db := config.NewMysqlDB(appConfig.MysqlConfig)

	// 初始化 Redis
	redisClient := config.NewRedisClient(appConfig.RedisConfig, ctx)

	// 初始化资源提供者
	resProvider := NewResourceProvider(redisClient, db)

	// 注册全局中间（局部中间件在具体的路由处注册）
	middleware.RegisterMiddleware(engine, db, redisClient)
	// 注册路由
	route.Register(engine, resProvider)

	// 获取启动端口
	// 定义一个命令行参数，默认值为 8080，描述为 "port to listen on"
	port := flag.Int("port", 8080, "port to listen on")
	// 解析命令行参数
	flag.Parse()

	log.Info("server listening on port", zap.Int("port", *port))

	// 创建一个HTTP服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: engine,
	}

	return &server{
		impl:        srv,
		resProvider: resProvider}
}
