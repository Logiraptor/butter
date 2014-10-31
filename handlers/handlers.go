package handlers

import (
	"net/http"

	"appengine"

	"github.com/Logiraptor/butter/keys"

	"code.google.com/p/go.net/context"
)

// A Response can encode itself onto the wire
type Response interface {
	Encode(http.ResponseWriter) error
}

// HandlerFunc wraps a function in order to turn it into a Handler
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) Response

func (h HandlerFunc) ServeHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request) Response {
	return h(ctx, rw, req)
}

// Handler is an extension of http.Handler which adds a context.
// See https://blog.golang.org/context for the rationale behind
// breaking the http.Handler convention
type Handler interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request) Response
}

// BaseFunc is a convenient wrapper around Base.
func BaseFunc(h func(context.Context, http.ResponseWriter, *http.Request) Response) http.Handler {
	return Base(HandlerFunc(h))
}

// Base converts a Handler into a regular http.Handler
func Base(h Handler) http.Handler {
	return base{h}
}

type base struct {
	h Handler
}

func (b base) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	err := b.h.ServeHTTP(context.Background(), rw, req).Encode(rw)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
	}
}

// AEContext wraps a handler and injects an appengine.Context
// only if one does not already exist
func AEContext(h Handler) Handler {
	return aeContext{h}
}

type aeContext struct {
	h Handler
}

func (a aeContext) ServeHTTP(ctx context.Context, rw http.ResponseWriter, req *http.Request) Response {
	if c := keys.AEContext(ctx); c != nil {
		return a.h.ServeHTTP(ctx, rw, req)
	}
	return a.h.ServeHTTP(keys.WithAEContext(ctx, appengine.NewContext(req)), rw, req)
}
