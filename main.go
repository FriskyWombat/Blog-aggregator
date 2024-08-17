package main

/*
"postgres://postgres:ada@localhost:5432/blogator"
*/

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/friskywombat/blog-aggregator/internal/database"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

type key struct{}

var userKey = key{}

func newConfig(dbURL string) apiConfig {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	return apiConfig{
		DB: database.New(db),
	}
}

func main() {
	godotenv.Load()
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DBCONN")
	cfg := newConfig(dbURL)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Mount("/v1", router)
	router.Get("/healthz", healthzHandleFunc)
	router.Get("/err", errorHandleFunc)
	router.Post("/users", cfg.newUserHandleFunc)
	router.Get("/users", cfg.authenticate(cfg.getUserHandleFunc))
	router.Post("/feeds", cfg.authenticate(cfg.newFeedHandleFunc))

	server := http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Fatal(server.ListenAndServe())
}

func (cfg *apiConfig) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := strings.Fields(r.Header.Get("Authorization"))
		if len(headers) < 2 || headers[0] != "ApiKey" {
			respondWithError(w, 401, "Unauthorized")
			return
		}
		user, err := cfg.DB.GetUser(r.Context(), headers[1])
		if err != nil {
			respondWithError(w, 400, "Bad authentication")
			return
		}
		ctxt := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctxt))
	})
}
