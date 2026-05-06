package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/resp"
	"go.uber.org/zap"
)

func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					zap.Any("error", err),
					zap.String("path", string(c.URI().Path())),
					zap.String("stack", string(debug.Stack())),
				)
				resp.InternalError(c, fmt.Sprintf("internal server error: %v", err))
				c.Abort()
			}
		}()
		c.Next(ctx)
	}
}
