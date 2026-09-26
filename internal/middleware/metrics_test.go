package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"txing-ai/internal/global"
	"txing-ai/internal/global/logging"
	obs "txing-ai/internal/observability"
)

// TestMetricsMiddleware 验证 HTTP RED 中间件计数（含 404 与 panic-500）
func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 初始化全局 logger（panic 路径的 customHandleRecovery 会用到），写到临时文件
	logging.InitLogger(&global.LogConfig{
		Level: "error", FileName: filepath.Join(t.TempDir(), "test.log"), MaxSize: 1,
	}, "test")

	cfg := global.ObservabilityConfig{Enabled: true}
	restore := obs.SetConfigForTest(&cfg)
	defer restore()
	obs.Setup(&cfg)

	engine := gin.New()
	engine.Use(Metrics())
	engine.Use(RecoveryWithZap(zap.NewNop(), false))
	engine.GET("/api/ping", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	engine.GET("/api/boom", func(c *gin.Context) { panic("boom") })

	do := func(path string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		engine.ServeHTTP(w, req)
		return w
	}

	if w := do("/api/ping"); w.Code != 200 {
		t.Fatalf("ping expect 200, got %d", w.Code)
	}
	// 项目约定：panic 被 Recovery 捕获后走 HTTP 200 + 错误码信封（非 500），
	// 指标按实际状态码记录
	if w := do("/api/boom"); w.Code != 200 {
		t.Fatalf("boom expect 200 (error envelope), got %d", w.Code)
	}
	if w := do("/api/notexist"); w.Code != 404 {
		t.Fatalf("404 expect 404, got %d", w.Code)
	}
	do("/static/ignored") // 非 /api 前缀不计

	obs.SampleNow() // 捕获本次增量

	now := time.Now()
	sums := obs.Summarize("txing_http_requests_total", "code", now.Add(-time.Minute), now.Add(time.Minute))
	byCode := map[string]float64{}
	for _, s := range sums {
		byCode[s.Key] = s.Sum
	}
	if byCode["200"] != 2 || byCode["404"] != 1 {
		t.Fatalf("expect 200x2/404x1, got %+v", byCode)
	}
}
