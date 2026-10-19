// Package server provides the HTTP server for the OVH exporter.
package server

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config holds server configuration.
type Config struct {
	MetricsPath string
	Port        int
}

// Server is the HTTP server for the exporter.
type Server struct {
	config   Config
	registry *prometheus.Registry
	readyFn  func() bool
}

// New creates a new Server.
func New(config Config, registry *prometheus.Registry, readyFn func() bool) *Server {
	return &Server{
		config:   config,
		registry: registry,
		readyFn:  readyFn,
	}
}

// Handler returns the HTTP handler with all routes configured.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleLanding)
	mux.HandleFunc("/-/healthy", s.handleHealthy)
	mux.HandleFunc("/-/ready", s.handleReady)
	mux.Handle(s.config.MetricsPath, promhttp.HandlerFor(s.registry, promhttp.HandlerOpts{}))

	return mux
}

// HTTPServer returns a configured *http.Server.
func (s *Server) HTTPServer() *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Port),
		Handler: s.Handler(),
	}
}

func (s *Server) handleLanding(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>OVH Exporter</title></head>
<body>
<h1>OVH Exporter</h1>
<p><a href="%s">Metrics</a></p>
</body>
</html>`, s.config.MetricsPath)
}

func (s *Server) handleHealthy(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request) {
	if !s.readyFn() {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "Not Ready")
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
