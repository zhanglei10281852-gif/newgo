package requestmeta

import "context"

type key string

const requestKey key = "request-id"

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestKey, id)
}
func ID(ctx context.Context) string { value, _ := ctx.Value(requestKey).(string); return value }
