package traefik_empty_gif_plugin_test

import (
	"context"
	"image/gif"
	"net/http"
	"net/http/httptest"
	"testing"

	plugin "github.com/suer/traefik-empty-gif-plugin"
)

func newHandler(t *testing.T) http.Handler {
	t.Helper()
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("next handler must not be called")
	})
	return newHandlerWithNext(t, next)
}

func newHandlerWithNext(t *testing.T, next http.Handler) http.Handler {
	t.Helper()
	cfg := plugin.CreateConfig()
	cfg.Paths = []string{"/favicon.ico", "/x"}
	h, err := plugin.New(context.Background(), next, cfg, "empty-gif")
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestNoPaths(t *testing.T) {
	_, err := plugin.New(context.Background(), http.NotFoundHandler(), plugin.CreateConfig(), "empty-gif")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPassThrough(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusTeapot)
	})
	for _, path := range []string{"/", "/other", "/favicon.ico/", "/x/y"} {
		rec := httptest.NewRecorder()
		newHandlerWithNext(t, next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusTeapot {
			t.Errorf("%s: status = %d", path, rec.Code)
		}
	}
}

func TestGet(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/gif" {
		t.Errorf("content-type = %q", ct)
	}
	if rec.Body.Len() != 43 {
		t.Errorf("body length = %d", rec.Body.Len())
	}
	img, err := gif.Decode(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 1 || b.Dy() != 1 {
		t.Errorf("size = %v", b)
	}
}

func TestHead(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler(t).ServeHTTP(rec, httptest.NewRequest(http.MethodHead, "/x", nil))

	if rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Errorf("status = %d, body = %d", rec.Code, rec.Body.Len())
	}
	if cl := rec.Header().Get("Content-Length"); cl != "43" {
		t.Errorf("content-length = %q", cl)
	}
}

func TestPost(t *testing.T) {
	rec := httptest.NewRecorder()
	newHandler(t).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/x", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d", rec.Code)
	}
}
