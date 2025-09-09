package logger

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func InitLogger(level string) {
	var cfg zap.Config

	env := "development"

	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	case "fatal":
		zapLevel = zapcore.FatalLevel
	default:
		zapLevel = zapcore.InfoLevel
	}
	cfg.Level.SetLevel(zapLevel)

	var err error
	log, err = cfg.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(log)
}

func L() *zap.Logger {
	return log
}

func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		ip := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if statusCode >= http.StatusInternalServerError {
			logger.Error("HTTP Access",
				zap.Int("status", statusCode),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", ip),
				zap.Duration("latency", latency),
				zap.String("user-agent", c.Request.UserAgent()),
				zap.String("error", errorMsg),
			)
		} else {
			logger.Info("HTTP Access",
				zap.Int("status", statusCode),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("query", query),
				zap.String("ip", ip),
				zap.Duration("latency", latency),
				zap.String("user-agent", c.Request.UserAgent()),
			)
		}
	}
}
