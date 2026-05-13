package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"typeform-test/config"
	"typeform-test/handlers"
	"typeform-test/store"
)

func main() {
	cfg := config.Load()
	s, err := store.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("could not initialize store: %v", err)
	}
	h := handlers.NewHandler(s)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /forms", h.CreateForm)
	mux.HandleFunc("POST /forms/{id}/questions", h.AddQuestion)
	mux.HandleFunc("POST /forms/{id}/save", h.SaveForm)
	mux.HandleFunc("GET /forms", h.ListForms)
	mux.HandleFunc("GET /forms/{id}", h.GetForm)
	mux.HandleFunc("GET /form/{slug}", h.GetFormBySlug)
	mux.HandleFunc("POST /form/{slug}/responses", h.SubmitResponse)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s - High Concurrency Mode!", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("could not start server: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
