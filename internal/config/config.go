package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/goczangabor24/chirpy/internal/auth"
	"github.com/goczangabor24/chirpy/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Db             *database.Queries
	Platform       string
	JWTSecret      string
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) Metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, fmt.Sprintf(
		`<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
		</html>`,
		int(cfg.FileserverHits.Load())))
}

func (cfg *ApiConfig) Reset(w http.ResponseWriter, _ *http.Request) {
	if cfg.Platform != "dev" {
		respondWithError(w, 403, "Forbidden")
		return
	}

	cfg.FileserverHits.Store(0)
	io.WriteString(w, fmt.Sprintln("Hits counter reset to 0, 'users' database entries deleted"))
	cfg.Db.DeleteUsers(context.Background())
}

func (cfg *ApiConfig) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK\n")
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		HashedPassword string `json:"password"`
		Email          string `json:"email"`
	}

	type ReturnVals struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameter: %s", err)
		w.WriteHeader(500)
		return
	}

	hashedPassword, err := auth.HashPassword(params.HashedPassword)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	createUserParams := database.CreateUserParams{
		ID:             uuid.New(),
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	response, err := cfg.Db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		respondWithError(w, 400, "User already exists")
		return
	}

	responseJSON := ReturnVals{
		ID:        response.ID,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.UpdatedAt,
		Email:     response.Email,
	}

	respondWithJSON(w, 201, responseJSON)
}

func (cfg *ApiConfig) Login(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	var params Parameters
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v", err)
		w.WriteHeader(500)
		return
	}

	response, err := cfg.Db.Login(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	correctPassword, err := auth.CheckPassword(params.Password, response.HashedPassword)

	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	} else if !correctPassword {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = 3600
	} else if params.ExpiresInSeconds > 3600 {
		params.ExpiresInSeconds = 3600
	}

	token, err := auth.MakeJWT(response.ID, cfg.JWTSecret, time.Second*time.Duration(params.ExpiresInSeconds))
	if err != nil {
		fmt.Println(err)
		return
	}

	type returnVals struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	resp := returnVals{
		ID:        response.ID,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.UpdatedAt,
		Email:     response.Email,
		Token:     token,
	}

	respondWithJSON(w, 200, resp)
}

func (cfg *ApiConfig) GetAllChirps(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	chirps, err := cfg.Db.GetAllChirps(r.Context())
	if err != nil {
		fmt.Println(err)
		return
	}

	response := []Chirp{}

	for _, chirp := range chirps {
		response = append(response, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.ID,
		})
	}

	dat, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Error marshaling response: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *ApiConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		w.WriteHeader(404)
		fmt.Print("Error parsing the UUID")
		return
	}

	chirp, err := cfg.Db.GetChirp(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		fmt.Printf("Chirp (ID: %v) doesn't exist in the database", chirpID)
		return
	}

	var response Chirp

	response = Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.ID,
	}

	dat, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *ApiConfig) Chirps(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		Body  string `json:"body"`
		Token string `json:"token"`
	}

	type ReturnVals struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("coding parameters: %s", err)
		w.WriteHeader(400)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	validID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "Message can't be longer than 140 characters")
		return
	}

	chirpParameters := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      replaceProfaneWords(params.Body),
		UserID:    validID,
	}

	response, err := cfg.Db.CreateChirp(r.Context(), chirpParameters)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintln(err))
		return
	}

	respBody := ReturnVals{
		ID:        response.ID,
		CreatedAt: response.CreatedAt,
		UpdatedAt: response.UpdatedAt,
		Body:      response.Body,
		UserID:    response.UserID,
	}

	respondWithJSON(w, 201, respBody)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type payload struct {
		Error string `json:"error"`
	}

	body := payload{Error: msg}

	dat, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(400)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func replaceProfaneWords(text string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	var cleanedText []string

	words := strings.Split(text, " ")

	for _, word := range words {
		if slices.Contains(profaneWords, strings.ToLower(word)) {
			cleanedText = append(cleanedText, "****")
		} else {
			cleanedText = append(cleanedText, word)
		}
	}

	return strings.Join(cleanedText, " ")
}
