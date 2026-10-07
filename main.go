package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/goczangabor24/chirpy/internal/config"
	"github.com/goczangabor24/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()

	srv := http.Server{
		Addr:         ":8080",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		Handler:      http.NewCrossOriginProtection().Handler(mux),
	}

	var apiCfg config.ApiConfig
	apiCfg.Db = dbQueries
	apiCfg.Platform = os.Getenv("PLATFORM")
	apiCfg.JWTSecret = os.Getenv("SECRET")

	fileserver := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	//HANDLERS
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(fileserver))
	mux.HandleFunc("GET /admin/healthz", apiCfg.Health)
	mux.HandleFunc("GET /admin/metrics", apiCfg.Metrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.Reset)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetAllChirps)
	mux.HandleFunc("POST /api/chirps", apiCfg.Chirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirp)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.Login)

	log.Println("Server running on http://localhost:8080")

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
