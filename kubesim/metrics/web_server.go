package metrics

import (
	"fmt"
	"html"
	"net/http"
	"sync"
)

// WebServer serves the latest TODO
type WebServer struct {
	latestNodesMetrics NodesMetrics
	latestPodsMetrics  PodsMetrics
	lock               sync.Mutex

	server *http.Server
}

// NewWebServer creates a new server that serves latest metrics.
func NewWebServer(port int) *WebServer {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: handler,
		// ReadTimeout:    10 * time.Second,
		// WriteTimeout:   10 * time.Second,
		// MaxHeaderBytes: 1 << 20,
	}

	return &WebServer{
		latestNodesMetrics: NodesMetrics{},
		latestPodsMetrics:  PodsMetrics{},
		lock:               sync.Mutex{},
		server:             server,
	}
}

func (w *WebServer) Write(nodesMetrics NodesMetrics, podsMetrics PodsMetrics) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.latestNodesMetrics = nodesMetrics
	w.latestPodsMetrics = podsMetrics

	return nil
}

func (w *WebServer) Serve() error {
	return w.server.ListenAndServe()
}

func (w *WebServer) Close() error {
	// TODO
	return nil
}

var _ = Writer(&WebServer{})
