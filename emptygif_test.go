package traefik_empty_gif_plugin_test

import (
	"context"
	"image/gif"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

func TestDefaultPath(t *testing.T) {
	h, err := plugin.New(context.Background(), http.NotFoundHandler(), plugin.CreateConfig(), "empty-gif")
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.ico", nil))
	if rec.Code != http.StatusOK || rec.Body.Len() != 43 {
		t.Errorf("/favicon.ico: status = %d, body = %d", rec.Code, rec.Body.Len())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/other", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("/other: status = %d", rec.Code)
	}
}

func TestNilConfig(t *testing.T) {
	if _, err := plugin.New(context.Background(), http.NotFoundHandler(), nil, "empty-gif"); err != nil {
		t.Fatal(err)
	}
}

func TestCacheControl(t *testing.T) {
	tests := []struct {
		expires string
		want    string
		wantErr bool
	}{
		{expires: "", want: "max-age=3600"},
		{expires: "off", want: ""},
		{expires: "1h", want: "max-age=3600"},
		{expires: "90s", want: "max-age=90"},
		{expires: "0s", want: "max-age=0"},
		{expires: "abc", wantErr: true},
		{expires: "-1h", wantErr: true},
	}
	for _, tt := range tests {
		cfg := plugin.CreateConfig()
		cfg.Expires = tt.expires
		h, err := plugin.New(context.Background(), http.NotFoundHandler(), cfg, "empty-gif")
		if tt.wantErr {
			if err == nil {
				t.Errorf("%q: expected error", tt.expires)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%q: %v", tt.expires, err)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.ico", nil))
		if got := rec.Header().Get("Cache-Control"); got != tt.want {
			t.Errorf("%q: Cache-Control = %q, want %q", tt.expires, got, tt.want)
		}

		expires := rec.Header().Get("Expires")
		if tt.want == "" {
			if expires != "" {
				t.Errorf("%q: Expires = %q, want none", tt.expires, expires)
			}
			continue
		}
		exp, err := http.ParseTime(expires)
		if err != nil {
			t.Fatalf("%q: Expires = %q: %v", tt.expires, expires, err)
		}
		d, _ := time.ParseDuration(strings.TrimPrefix(tt.want, "max-age=") + "s")
		if diff := time.Until(exp) - d; diff > 2*time.Second || diff < -2*time.Second {
			t.Errorf("%q: Expires is off by %v", tt.expires, diff)
		}
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
