package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/zy84338719/ikuai-tools-service/internal/pkg/ctxkey"
	log "github.com/zy84338719/ikuai-tools-service/internal/pkg/logger"
	"go.uber.org/zap"
)

const headerRequestID = "X-Request-ID"

func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.URI().Path())
		method := string(c.Method())

		// Attach or propagate a request ID for log correlation.
		reqID := string(c.GetHeader(headerRequestID))
		if reqID == "" {
			reqID = uuid.New().String()
		}
		c.Header(headerRequestID, reqID)

		// Inject the request id into the request context so downstream
		// handlers and jobs can correlate their logs (ctxkey.RequestID).
		// Passing the enriched ctx to c.Next propagates it to handlers.
		ctx = ctxkey.WithRequestID(ctx, reqID)

		defer func() {
			log.Info("HTTP",
				zap.String("id", reqID),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("ip", c.ClientIP()),
				zap.Int("status", c.Response.StatusCode()),
				zap.Duration("latency", time.Since(start)),
			)
		}()

		c.Next(ctx)
	}
}
