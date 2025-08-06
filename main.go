package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ehumba/chirpy-web-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func handlerEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	// set up database
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	if err != nil {
		fmt.Printf("could not load database: %v", err)
		return
	}
	mux := http.NewServeMux()

	// Serve static files from the current directory
	handler := http.StripPrefix("/app", http.FileServer(http.Dir("app")))

	apiCfg := apiConfig{
		dbQueries: dbQueries,
		platform:  platform,
	}

	// Handle the root path
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /api/healthz", handlerEndpoint)

	// Counter endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerCount)

	// Reset endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	// New user creation endpoint
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	// Chirp creation endpoint
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirps)

	// Get Chirps endpoint
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)

	// Get Chirp endpoint
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)

	// Login endpoint
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) handlerCount(w http.ResponseWriter, r *http.Request) {
	hits := fmt.Sprintf(
		`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, a.fileserverHits.Load())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(hits))
}

func (a *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if a.platform != "dev" {
		respondWithError(w, 403, "you don't have access to this endpoint")
		return
	}
	a.fileserverHits.Store(0)
	err := a.dbQueries.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, "failed to delete user data")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Reset"))
}
