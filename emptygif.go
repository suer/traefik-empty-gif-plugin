// Package traefik_empty_gif_plugin answers every request with a 1x1 transparent GIF.
package traefik_empty_gif_plugin

import (
	"context"
	"errors"
	"net/http"
	"strconv"
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

// Config is the plugin configuration.
type Config struct {
	// Paths lists the exact request paths answered with the GIF.
	Paths []string `json:"paths,omitempty" yaml:"paths,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type emptyGIFHandler struct {
	next  http.Handler
	paths map[string]struct{}
}

// New creates the middleware.
func New(_ context.Context, next http.Handler, config *Config, _ string) (http.Handler, error) {
	if config == nil || len(config.Paths) == 0 {
		return nil, errors.New("paths must not be empty")
	}

	paths := make(map[string]struct{}, len(config.Paths))
	for _, p := range config.Paths {
		paths[p] = struct{}{}
	}

	return &emptyGIFHandler{next: next, paths: paths}, nil
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
	rw.WriteHeader(http.StatusOK)

	if req.Method == http.MethodGet {
		_, _ = rw.Write(emptyGIF)
	}
}
