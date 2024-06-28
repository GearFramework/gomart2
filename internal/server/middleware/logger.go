package middleware

import (
	"github.com/GearFramework/gomart2/internal/pkg/alog"
	"github.com/gin-gonic/gin"
	"time"
)

func Logger() gin.HandlerFunc {
	logger := alog.NewLogger("info")
	return func(ctx *gin.Context) {
		start := time.Now()
		logger.Infof("%s request: %s",
			ctx.Request.Method,
			ctx.Request.RequestURI,
		)
		ctx.Next()
		duration := getDurationInMilliseconds(start)
		logger.Infof("%s response from: %s; status: %d; size: %d | duration: %.4f ms",
			ctx.Request.Method,
			ctx.Request.RequestURI,
			ctx.Writer.Status(),
			ctx.Writer.Size(),
			duration,
		)
	}
}

func getDurationInMilliseconds(start time.Time) float64 {
	end := time.Now()
	duration := end.Sub(start)
	milliseconds := float64(duration) / float64(time.Millisecond)
	rounded := float64(int(milliseconds*100+.5)) / 100
	return rounded
}
