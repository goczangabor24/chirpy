package config

import (
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) Metrics(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, fmt.Sprintf("Hits: %d", int(cfg.fileserverHits.Load())))
}

func (cfg *ApiConfig) Reset(w http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits.Store(0)
	io.WriteString(w, fmt.Sprintln("Hits counter reset to 0"))
}
