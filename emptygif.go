// Package traefik_empty_gif_plugin answers every request with a 1x1 transparent GIF.
package traefik_empty_gif_plugin

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Same 43 bytes that nginx ngx_http_empty_gif_module serves.
var emptyGIF = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x01,
	0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x21, 0xf9, 0x04, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
	0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
}

// nginx uses a fixed Last-Modified for this module.
const lastModified = "Mon, 28 Sep 1970 06:00:00 GMT"

const (
	defaultPath    = "/favicon.ico"
	defaultExpires = "1h"
	expiresOff     = "off"
)

// Config is the plugin configuration.
type Config struct {
	// Paths lists the exact request paths answered with the GIF.
	// Defaults to ["/favicon.ico"].
	Paths []string `json:"paths,omitempty" yaml:"paths,omitempty"`
	// Expires is a Go duration (e.g. "1h") or "off", like the nginx expires directive:
	// it sets Expires and Cache-Control: max-age. Defaults to "1h".
	Expires string `json:"expires,omitempty" yaml:"expires,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type emptyGIFHandler struct {
	next         http.Handler
	paths        map[string]struct{}
	cacheControl string
	maxAge       time.Duration
	setExpires   bool
}

// New creates the middleware.
func New(_ context.Context, next http.Handler, config *Config, _ string) (http.Handler, error) {
	if config == nil {
		config = CreateConfig()
	}

	configured := config.Paths
	if len(configured) == 0 {
		configured = []string{defaultPath}
	}
	paths := make(map[string]struct{}, len(configured))
	for _, p := range configured {
		paths[p] = struct{}{}
	}

	handler := &emptyGIFHandler{next: next, paths: paths}
	expires := config.Expires
	if expires == "" {
		expires = defaultExpires
	}
	if expires != expiresOff {
		d, err := time.ParseDuration(expires)
		if err != nil {
			return nil, fmt.Errorf("invalid expires: %w", err)
		}
		if d < 0 {
			return nil, fmt.Errorf("invalid expires: %q must not be negative", expires)
		}
		handler.setExpires = true
		handler.maxAge = d.Truncate(time.Second)
		handler.cacheControl = "max-age=" + strconv.FormatInt(int64(handler.maxAge/time.Second), 10)
	}

	return handler, nil
}

func (h *emptyGIFHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if _, ok := h.paths[req.URL.Path]; !ok {
		h.next.ServeHTTP(rw, req)
		return
	}

	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		rw.Header().Set("Allow", "GET, HEAD")
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rw.Header().Set("Content-Type", "image/gif")
	rw.Header().Set("Content-Length", strconv.Itoa(len(emptyGIF)))
	rw.Header().Set("Last-Modified", lastModified)
	if h.setExpires {
		rw.Header().Set("Expires", time.Now().Add(h.maxAge).UTC().Format(http.TimeFormat))
		rw.Header().Set("Cache-Control", h.cacheControl)
	}
	rw.WriteHeader(http.StatusOK)

	if req.Method == http.MethodGet {
		_, _ = rw.Write(emptyGIF)
	}
}
