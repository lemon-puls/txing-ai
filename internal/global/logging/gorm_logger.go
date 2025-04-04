package logging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// GormLogger 实现了 GORM 的日志接口，用于将 GORM 的日志输出到 Zap
type GormLogger struct {
	ZapLogger                 *zap.Logger     // Zap 日志实例
	LogLevel                  logger.LogLevel // 日志级别
	SlowThreshold             time.Duration   // 慢查询阈值
	IgnoreRecordNotFoundError bool            // 是否忽略记录未找到的错误
}

// NewGormLogger 创建一个新的 GORM 日志记录器实例
func NewGormLogger(zapLogger *zap.Logger) *GormLogger {
	return &GormLogger{
		ZapLogger:                 zapLogger,
		LogLevel:                  logger.Info, // 默认日志级别为 Info
		SlowThreshold:             time.Second, // 默认慢查询阈值为 1 秒
		IgnoreRecordNotFoundError: true,        // 默认忽略记录未找到的错误
	}
}

// LogMode 设置日志级别并返回新的日志记录器实例
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 记录 Info 级别的日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.ZapLogger.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn 记录 Warn 级别的日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.ZapLogger.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error 记录 Error 级别的日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.ZapLogger.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace 记录 SQL 执行的跟踪信息
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	// 如果日志级别为 Silent，直接返回
	if l.LogLevel <= logger.Silent {
		return
	}

	// 计算执行耗时
	elapsed := time.Since(begin)
	// 获取 SQL 语句和影响的行数
	sql, rows := fc()

	// 构建日志字段
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	// 根据不同情况记录日志
	switch {
	// 发生错误且不忽略记录未找到错误时，记录错误日志
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		l.ZapLogger.Error("gorm error", append(fields, zap.Error(err))...)
	// 执行时间超过慢查询阈值时，记录警告日志
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		l.ZapLogger.Warn("gorm slow sql", fields...)
	// 其他情况记录普通的追踪日志
	case l.LogLevel >= logger.Info:
		l.ZapLogger.Info("gorm trace", fields...)
	}
}
