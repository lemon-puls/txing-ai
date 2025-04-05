package log

import (
	"txing-ai/internal/global/logging"
)

// 导出便捷的日志方法
// Export convenient logging methods
var (
	Debug  = logging.Debug
	Info   = logging.Info
	Warn   = logging.Warn
	Error  = logging.Error
	Fatal  = logging.Fatal
	With   = logging.With
	Named  = logging.Named
	Logger = logging.Logger
)
