package kami // import "gopkg.in/thinxer/semikami.v2"

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

type (
	// FilterFunc runs before the actual handler. The pipeline gets canceled when nil is returned.
	FilterFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context
	// WrapFunc is more like a traditional middleware, where it calls next to continue the execution.
	WrapFunc func(next HandlerFunc) HandlerFunc
	// HandlerFunc is at the end of the execution.
	HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)
)

// Builder provides fluent APIs for a minimal middleware framework.
type Builder struct {
	router  *httprouter.Router
	pre     WrapFunc
	befores []FilterFunc
}

// New creates a Builder. All builders derived from the same root builder shares the same router.
func New(router *httprouter.Router) Builder {
	if router == nil {
		router = httprouter.New()
	}
	return Builder{
		router: router,
		pre: func(next HandlerFunc) HandlerFunc {
			return next
		},
	}
}

func (b *Builder) runthrough(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	for _, m := range b.befores {
		ctx = m(ctx, w, r)
		if ctx == nil {
			break
		}
	}
	return ctx
}

// ServeHTTP implements http.Handler interface.
func (b Builder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.router.ServeHTTP(w, r)
}

// With creates a Builder by copying the parent Builder and appending middleware m.
func (b Builder) With(m FilterFunc) Builder {
	n := len(b.befores)
	return Builder{
		router:  b.router,
		pre:     b.pre,
		befores: append(b.befores[:n:n], m),
	}
}

// Wrap creates a Builder by copying the parent Builder and appending wrapper f.
func (b Builder) Wrap(f WrapFunc) Builder {
	return Builder{
		router: b.router,
		pre: func(next HandlerFunc) HandlerFunc {
			wrapped := f(next)
			return b.pre(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				if ctx = b.runthrough(ctx, w, r); ctx == nil {
					return
				}
				wrapped(ctx, w, r)
			})
		},
	}
}

// Handle creates the router entry for h.
func (b Builder) Handle(method string, path string, h HandlerFunc) {
	wrapped := b.pre(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if ctx = b.runthrough(ctx, w, r); ctx == nil {
			return
		}
		h(ctx, w, r)
	})
	b.router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, paramKey, ps)
		wrapped(ctx, w, r)
	})
}

// Get creates router entry for h with GET method.
func (b Builder) Get(path string, h HandlerFunc) { b.Handle("GET", path, h) }

// Post creates router entry for h with POST method.
func (b Builder) Post(path string, h HandlerFunc) { b.Handle("POST", path, h) }

// Delete creates router entry for h with DELETE method.
func (b Builder) Delete(path string, h HandlerFunc) { b.Handle("DELETE", path, h) }

// Patch creates router entry for h with PATCH method.
func (b Builder) Patch(path string, h HandlerFunc) { b.Handle("PATCH", path, h) }

// Head creates router entry for h with HEAD method.
func (b Builder) Head(path string, h HandlerFunc) { b.Handle("HEAD", path, h) }

// Options creates router entry for h with OPTIONS method.
func (b Builder) Options(path string, h HandlerFunc) { b.Handle("OPTIONS", path, h) }

type contextKeyType int

const (
	paramKey contextKeyType = iota
)

// Param extracts the named route param from the context.
func Param(ctx context.Context, name string) string {
	params, ok := ctx.Value(paramKey).(httprouter.Params)
	if !ok {
		panic("semikami: httprouter.Params not available in this context.")
	}
	return params.ByName(name)
}
