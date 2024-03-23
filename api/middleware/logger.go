package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const logKey = "logger"

func Logger(logger logrus.FieldLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func(start time.Time) {
			statusCode := ctx.Writer.Status()

			loggerContext := GetLoggerContext(ctx)
			log := logger.
				WithFields(loggerContext).
				WithFields(logrus.Fields{
					"latency":    time.Since(start).String(),
					"statusCode": statusCode,
					"reqParams":  ctx.Request.URL.Query(),
				})

			switch {
			case statusCode >= http.StatusInternalServerError:
				log.Error(ctx.Errors)
			case statusCode >= http.StatusBadRequest:
				log.Warn(ctx.Errors)
			default:
				log.Info("OK")
			}
		}(time.Now())

		SetLoggerContext(ctx)
		ctx.Next()
	}
}

func SetLoggerContext(ctx *gin.Context) {
	logContext := logrus.Fields{
		"clientIP":    ctx.ClientIP(),
		"method":      ctx.Request.Method,
		"path":        ctx.Request.URL.Path,
		"escapedPath": ctx.Request.URL.EscapedPath(),
		"handler":     ctx.HandlerName(),
	}

	ctx.Set(logKey, logContext)
}

func GetLoggerContext(ctx *gin.Context) logrus.Fields {
	logContext, ok := ctx.Value(logKey).(logrus.Fields)
	if !ok {
		logContext = logrus.Fields{}
	}

	return logContext
}
