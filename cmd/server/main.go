package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/tahoorian/tahoorian/internal/config"
	"github.com/tahoorian/tahoorian/internal/handler"
	"github.com/tahoorian/tahoorian/internal/middleware"
	memrepo "github.com/tahoorian/tahoorian/internal/repository/memory"
	"github.com/tahoorian/tahoorian/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	repo := memrepo.New()
	svc := service.New(repo)
	h := handler.New(svc)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/", h.Index)
	r.Get("/about", h.About)
	r.Get("/services", h.Services)
	r.Get("/projects", h.Projects)
	r.Get("/team", h.Team)
	r.Get("/contact", h.Contact)
	r.Post("/api/contact", h.SubmitContact)

	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static")))
	r.Get("/static/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		fileServer.ServeHTTP(w, r)
	}).ServeHTTP)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	slog.Info("starting server", "port", cfg.Port, "env", cfg.Env)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
