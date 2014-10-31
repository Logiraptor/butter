package keys

import (
	"code.google.com/p/go.net/context"

	"appengine"
)

type key int

var (
	// aeContext is the conventional key inside butter for an appengine Context
	aeContext key = 1
)

// WithAEContext adds an appengine context to the base context.
func WithAEContext(ctx context.Context, c appengine.Context) context.Context {
	return context.WithValue(ctx, aeContext, c)
}

// AEContext returns the appengine context inside ctx, or nil if none exists
func AEContext(ctx context.Context) appengine.Context {
	c := ctx.Value(aeContext)
	if c == nil {
		return nil
	}
	return c.(appengine.Context)
}
