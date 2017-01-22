package kami

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrap(t *testing.T) {
	recorder := httptest.NewRecorder()
	kami := New(nil).With(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		return context.WithValue(ctx, 1, 2)
	}).Wrap(func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			if ctx.Value(1) == nil {
				t.Fatal("sequence error")
			}
			next(context.WithValue(ctx, 2, 3), w, r)
		}
	}).With(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
		if ctx.Value(2) == nil {
			t.Fatal("sequence error")
		}
		return context.WithValue(ctx, 3, 5)
	})
	kami.Get("/:key", func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if expected, got := "key", Param(ctx, "key"); expected != got {
			t.Fatal("expcted %v, got %v", expected, got)
		}
		v1 := ctx.Value(1).(int)
		v2 := ctx.Value(2).(int)
		v3 := ctx.Value(3).(int)
		w.WriteHeader(v1 * v2 * v3)
	})
	r, _ := http.NewRequest("GET", "/key", nil)
	kami.ServeHTTP(recorder, r)
	if expected, got := 2*3*5, recorder.Code; expected != got {
		t.Fatalf("expected %d, got %d", expected, got)
	}
}
