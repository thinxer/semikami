package kami

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

type (
	// BeforeFunc runs before the actual handler. The pipeline gets canceled when nil is returned.
	BeforeFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context
	// WrapFunc is more like a traditional middleware, where it calls next to continue the execution.
	WrapFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, next HandlerFunc)
	// HandlerFunc is at the end of the execution.
	HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)
	// Middleware is the old name of BeforeFunc.
	Middleware BeforeFunc
)

type Builder struct {
	router  *httprouter.Router
	pre     WrapFunc
	befores []BeforeFunc
}

// New creates a Kami Builder. All builders derived from the same root builder shares the same router.
func New(router *httprouter.Router) Builder {
	if router == nil {
		router = httprouter.New()
	}
	return Builder{
		router: router,
		pre: func(ctx context.Context, w http.ResponseWriter, r *http.Request, next HandlerFunc) {
			next(ctx, w, r)
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
func (b Builder) With(m BeforeFunc) Builder {
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
		pre: func(ctx context.Context, w http.ResponseWriter, r *http.Request, next HandlerFunc) {
			b.pre(ctx, w, r, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
				if ctx = b.runthrough(ctx, w, r); ctx == nil {
					return
				}
				f(ctx, w, r, next)
			})
		},
	}
}

func (b Builder) Handle(method string, path string, h HandlerFunc) {
	b.router.Handle(method, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, paramKey, ps)
		b.pre(ctx, w, r, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if ctx = b.runthrough(ctx, w, r); ctx == nil {
				return
			}
			h(ctx, w, r)
		})
	})
}

func (b Builder) Get(path string, h HandlerFunc) {
	b.Handle("GET", path, h)
}

func (b Builder) Post(path string, h HandlerFunc) {
	b.Handle("POST", path, h)
}
func (b Builder) Put(path string, h HandlerFunc) {
	b.Handle("PUT", path, h)
}

func (b Builder) Delete(path string, h HandlerFunc) {
	b.Handle("DELETE", path, h)
}

func (b Builder) Patch(path string, h HandlerFunc) {
	b.Handle("PATCH", path, h)
}

func (b Builder) Head(path string, h HandlerFunc) {
	b.Handle("HEAD", path, h)
}

func (b Builder) Options(path string, h HandlerFunc) {
	b.Handle("OPTIONS", path, h)
}

type contextKeyType int

const (
	paramKey contextKeyType = iota
)

func Param(ctx context.Context, name string) string {
	params, ok := ctx.Value(paramKey).(httprouter.Params)
	if !ok {
		panic("semikami: httprouter.Params not available in this context.")
	}
	return params.ByName(name)
}
