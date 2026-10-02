package main

import (
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

	//HANDLERS
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(fileserver))
	mux.HandleFunc("GET /admin/healthz", apiCfg.Health)
	mux.HandleFunc("GET /admin/metrics", apiCfg.Metrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.Reset)

	log.Println("Server running on http://localhost:8080")

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
