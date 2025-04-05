package logging

import (
	"os"
	"txing-ai/internal/global"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 全局logger实例
// Global logger instance
var Logger *zap.Logger

// InitLogger 初始化logger
// Initialize logger
func InitLogger(logConfig *global.LogConfig, profile string) {
	writeSyncer := getLogWriter(logConfig.FileName, logConfig.MaxSize, logConfig.MaxBackups, logConfig.MaxAge)
	encoder := getProdEncoder()
	var l = new(zapcore.Level)
	err := l.UnmarshalText([]byte(logConfig.Level))
	if err != nil {
		panic(err)
	}
	var core zapcore.Core
	if profile == "dev" {
		// 进入开发模式，日志输出到终端=
		consoleEncoder := getDevEncoder()
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, writeSyncer, l),
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
		)
	} else {
		core = zapcore.NewCore(encoder, writeSyncer, l)
	}

	Logger = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(Logger)

	Info("logger init success")
}

// Debug logs a message at DebugLevel
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal logs a message at FatalLevel
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// With creates a child logger with the given fields
func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}

// Named creates a child logger with the given name
func Named(name string) *zap.Logger {
	return Logger.Named(name)
}

func getProdEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getDevEncoder() zapcore.Encoder {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}
	return zapcore.AddSync(lumberJackLogger)
}
