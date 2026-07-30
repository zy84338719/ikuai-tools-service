package ctxkey

// Package ctxkey defines shared context keys used across middleware/handlers
// to pass per-request values (actor id, request id, ...) without import cycles.
import "context"

type key int

const (
	// actorKey holds the authenticated actor identity (api-key id / username /
	// "anonymous") for audit logging.
	actorKey key = iota
	// requestIDKey holds the X-Request-ID for log correlation.
	requestIDKey
)

// WithActor stores the actor identity in ctx.
func WithActor(ctx context.Context, actor string) context.Context {
	return context.WithValue(ctx, actorKey, actor)
}

// Actor retrieves the actor identity, or "anonymous" if unset.
func Actor(ctx context.Context) string {
	if v, ok := ctx.Value(actorKey).(string); ok && v != "" {
		return v
	}
	return "anonymous"
}

// WithRequestID stores the request id in ctx.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID retrieves the request id, or "" if unset.
func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}
