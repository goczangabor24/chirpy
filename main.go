package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/goczangabor24/chirpy/internal/config"
)

func main() {
	mux := http.NewServeMux()

	srv := http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		Handler:      http.NewCrossOriginProtection().Handler(mux),
	}

	var apiCfg config.ApiConfig

	fileserver := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(fileserver))

	readinessEndpoint := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK\n")
	}

	mux.HandleFunc("GET /healthz", readinessEndpoint)

	mux.HandleFunc("GET /metrics", apiCfg.Metrics)

	mux.HandleFunc("POST /reset", apiCfg.Reset)

	log.Println("Server running on http://localhost:8080")

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
