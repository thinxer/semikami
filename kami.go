package kami

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"golang.org/x/net/context"
)

type (
	Middleware  func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context
	HandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)
)

type Builder struct {
	background  func() context.Context
	middlewares []Middleware
	router      *httprouter.Router
}

func (b Builder) Handler() http.Handler {
	return http.HandlerFunc(b.router.ServeHTTP)
}

// With creates a Builder by copying the parent Builder and appending middleware m.
// The newly built Builder shares the router and background Context with its parent.
func (b Builder) With(m Middleware) Builder {
	ms := make([]Middleware, len(b.middlewares)+1)
	copy(ms, b.middlewares)
	ms[len(ms)-1] = m
	return Builder{
		background:  b.background,
		middlewares: ms,
		router:      b.router,
	}
}

func (b Builder) Get(path string, h HandlerFunc) {
	b.router.GET(path, wrap(b.background, b.middlewares, h))
}

func (b Builder) Post(path string, h HandlerFunc) {
	b.router.POST(path, wrap(b.background, b.middlewares, h))
}
func (b Builder) Put(path string, h HandlerFunc) {
	b.router.PUT(path, wrap(b.background, b.middlewares, h))
}

func (b Builder) Delete(path string, h HandlerFunc) {
	b.router.DELETE(path, wrap(b.background, b.middlewares, h))
}

func (b Builder) Patch(path string, h HandlerFunc) {
	b.router.PATCH(path, wrap(b.background, b.middlewares, h))
}

func (b Builder) Head(path string, h HandlerFunc) {
	b.router.HEAD(path, wrap(b.background, b.middlewares, h))
}

func (b Builder) Options(path string, h HandlerFunc) {
	b.router.OPTIONS(path, wrap(b.background, b.middlewares, h))
}

func (b Builder) Handle(method string, path string, h HandlerFunc) {
	b.router.Handle(method, path, wrap(b.background, b.middlewares, h))
}

func New(background func() context.Context, router *httprouter.Router) Builder {
	if background == nil {
		background = context.Background
	}
	if router == nil {
		router = httprouter.New()
	}
	return Builder{
		background: background,
		router:     router,
	}
}

func With(m Middleware) Builder {
	return New(nil, nil).With(m)
}

type contextKeyType int

const (
	paramKey contextKeyType = iota
)

func wrap(bg func() context.Context, middlewares []Middleware, h HandlerFunc) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ctx := context.WithValue(bg(), paramKey, ps)
		for _, m := range middlewares {
			ctx = m(ctx, w, r)
			if ctx == nil {
				return
			}
		}
		h(ctx, w, r)
	}
}

func Param(ctx context.Context, name string) string {
	params, ok := ctx.Value(paramKey).(httprouter.Params)
	if !ok {
		panic("semikami: httprouter.Params not available in this context.")
	}
	return params.ByName(name)
}
