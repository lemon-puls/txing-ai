package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
	"txing-ai/internal/global"
	"txing-ai/internal/utils"
)

type Fn func(c *gin.Context) []zapcore.Field

// Config is config setting for Ginzap
type Config struct {
	TimeFormat string
	UTC        bool
	SkipPaths  []string
	Context    Fn
}

// LoggerWithZap returns a gin.HandlerFunc (middleware) that logs requests using uber-go/zap.
//
// Requests with errors are logged using zap.Error().
// Requests without errors are logged using zap.Info().
//
// It receives:
//  1. A time package format string (e.g. time.RFC3339).
//  2. A boolean stating whether to use UTC time zone or local.
func LoggerWithZap(logger *zap.Logger, timeFormat string, utc bool) gin.HandlerFunc {
	return GinzapWithConfig(logger, &Config{TimeFormat: timeFormat, UTC: utc})
}

// GinzapWithConfig returns a gin.HandlerFunc using configs
func GinzapWithConfig(logger *zap.Logger, conf *Config) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		if _, ok := skipPaths[path]; !ok {
			end := time.Now()
			latency := end.Sub(start)
			if conf.UTC {
				end = end.UTC()
			}

			fields := []zapcore.Field{
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				//zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", c.ClientIP()),
				//zap.String("user-agent", c.Request.UserAgent()),
				zap.Duration("latency", latency),
			}
			if conf.TimeFormat != "" {
				//fields = append(fields, zap.String("time", end.Format(conf.TimeFormat)))
			}

			if conf.Context != nil {
				fields = append(fields, conf.Context(c)...)
			}

			if len(c.Errors) > 0 {
				// Append error field if this is an erroneous request.
				for _, e := range c.Errors.Errors() {
					logger.Error(e, fields...)
				}
			} else {
				logger.Info(path, fields...)
			}
		}
	}
}

func defaultHandleRecovery(c *gin.Context, err interface{}) {
	c.AbortWithStatus(http.StatusInternalServerError)
}

func customHandleRecovery(c *gin.Context, err interface{}) {
	switch v := err.(type) {
	case global.CustomError:
		utils.ErrorWithCode(c, v.Code, nil)
	case error:
		utils.ErrorWithMsg(c, v.Error(), v)
	case string:
		utils.ErrorWithMsg(c, v, nil)
	default:
		utils.ErrorWithMsg(c, "unknown error", nil)
	}
	c.Abort()
}

// RecoveryWithZap returns a gin.HandlerFunc (middleware)
// that recovers from any panics and logs requests using uber-go/zap.
// All errors are logged using zap.Error().
// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func RecoveryWithZap(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return CustomRecoveryWithZap(logger, stack, customHandleRecovery)
}

// CustomRecoveryWithZap returns a gin.HandlerFunc (middleware) with a custom recovery handler
// that recovers from any panics and logs requests using uber-go/zap.
// All errors are logged using zap.Error().
// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func CustomRecoveryWithZap(logger *zap.Logger, stack bool, recovery gin.RecoveryFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			path := c.Request.URL.Path
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.Error(path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				fields := []zapcore.Field{
					zap.Any("error", err),
					zap.String("path", path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
					zap.String("user-agent", c.Request.UserAgent()),
					//zap.String("request", string(httpRequest)),
				}

				if stack {
					fields = append(fields, zap.String("stack", string(debug.Stack())))
				}

				logger.Error("[Recovery from panic]", fields...)
				recovery(c, err)
			}
		}()
		c.Next()
	}
}
