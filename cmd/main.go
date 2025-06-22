package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"txing-ai/internal/app"
	"txing-ai/internal/global"
	"txing-ai/internal/global/config"
	"txing-ai/internal/global/logging"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
)

//go:generate swag init -g main.go -o ../docs

// @title        Txing AI API
// @version      1.0
// @description  Txing AI API文档
// @BasePath     /
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
// @host      localhost:8080
// @schemes   http
func main() {

	appConfig := global.LoadConfig()

	logging.InitLogger(appConfig.LogConfig, appConfig.Profile)

	log.Info("Txing AI service starting...")

	// 初始化翻译器，用于将错误信息等翻译成中文
	// 如果初始化失败，将会记录错误信息
	if err := config.InitTranslator("zh"); err != nil {
		log.Error("init translator failed", zap.Error(err))
	}

	// 监听中断信号（ctrl+c）和终止信号（kill -15 pid），用于实现服务的优雅关闭
	// 注意：kill -9（强制终止）信号无法被捕获，只能强制关闭服务
	sigCtx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// 启动服务
	app := app.New(sigCtx, appConfig)
	// 启动
	app.Start()

	// 打印服务访问地址
	log.Info("🚀 Txing AI service started successfully!")
	log.Info("\n _______     _                           _____ \n|__   __|   (_)                   /\\    |_   _|\n   | |__  __ _  _ __    __ _     /  \\     | |  \n   | |\\ \\/ /| || '_ \\  / _` |   / /\\ \\    | |  \n   | | >  < | || | | || (_| |  / ____ \\  _| |_ \n   |_|/_/\\_\\|_||_| |_| \\__, | /_/    \\_\\|_____|\n                        __/ |                  \n                       |___/   ")
	log.Info(fmt.Sprintf("📚 API Documentation: http://localhost:%d/swagger/index.html", appConfig.ServerConfig.Port))
	log.Info(fmt.Sprintf("🏠 Homepage: http://localhost:%d/dash", appConfig.ServerConfig.Port))
	log.Info("💡 Press Ctrl+C to stop the service")

	// 等待信号
	<-sigCtx.Done()

	// 创建一个超时时间为 5s 的 context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 调用 shutdown 方法，优雅关闭服务
	if err := app.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown:", zap.Error(err))
	}

	log.Info("Txing AI service stopped.")
}
