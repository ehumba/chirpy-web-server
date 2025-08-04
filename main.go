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

func endpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	// set up database
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
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
	}

	// Handle the root path
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	// Handle the endpoint
	mux.HandleFunc("GET /api/healthz", endpointHandler)

	// Handle the counter
	mux.HandleFunc("GET /admin/metrics", apiCfg.countHandler)

	// Handle the reset
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	// Handle the POST request
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.postReqHandler)

	// Handle new user creation
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) countHandler(w http.ResponseWriter, r *http.Request) {
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

func (a *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	a.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Reset"))
}
