package static

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var frontend embed.FS

func Register(router gin.IRouter) {
	// 获取dist子文件系统
	dist, err := fs.Sub(frontend, "frontend/dist")
	if err != nil {
		panic(err)
	}
	// 静态文件服务
	router.StaticFS("/", http.FS(dist))
}
