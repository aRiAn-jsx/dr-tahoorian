package main

import (
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/tahoorian/tahoorian/internal/config"
	"github.com/tahoorian/tahoorian/internal/handler"
	"github.com/tahoorian/tahoorian/internal/middleware"
	memrepo "github.com/tahoorian/tahoorian/internal/repository/memory"
	sqliterepo "github.com/tahoorian/tahoorian/internal/repository/sqlite"
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

	// --- Repository selection ---
	// SQLite is the production database. Set DB_TYPE=memory for ephemeral
	// in-memory storage (convenient for development / testing).
	var repo service.Repository
	switch strings.ToLower(cfg.DBType) {
	case "sqlite":
		dsn := cfg.DBDSN
		if dsn == "" {
			dsn = "tahoorian.db"
		}
		db, err := sql.Open("sqlite", dsn)
		if err != nil {
			slog.Error("failed to open sqlite database", "dsn", dsn, "error", err)
			os.Exit(1)
		}
		if err := sqliterepo.Migrate(db); err != nil {
			slog.Error("database migration failed", "error", err)
			os.Exit(1)
		}
		if err := sqliterepo.Seed(db); err != nil {
			slog.Warn("database seed skipped", "error", err)
		}
		repo = sqliterepo.NewRepository(db)
		slog.Info("using sqlite repository", "dsn", dsn)
	default:
		repo = memrepo.New()
		slog.Info("using in-memory repository")
	}

	// --- Session manager ---
	secret := cfg.SessionSecret
	if secret == "" {
		secret = "change-me-in-production"
		slog.Warn("SESSION_SECRET not set, using default — change this in production")
	}
	sm := middleware.NewSessionManager(secret)

	// --- Dev mode (re-parse templates on every request when ENV=development) ---
	devMode := strings.ToLower(cfg.Env) == "development"

	// --- Handlers ---
	svc := service.New(repo)
	h := handler.New(svc, devMode)
	adminH := handler.NewAdmin(svc, sm, devMode)

	// --- Router ---
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// Single CORS middleware (go-chi/cors is comprehensive; custom one removed)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// --- Public routes ---
	r.Get("/", h.Index)
	r.Get("/about", h.About)
	r.Get("/services", h.Services)
	r.Get("/projects", h.Projects)
	r.Get("/articles", h.Articles)
	r.Get("/articles/{slug}", h.ArticleDetail)
	r.Get("/team", h.Team)
	r.Get("/contact", h.Contact)
	r.Post("/api/contact", h.SubmitContact)

	// --- Admin routes (unauthenticated) ---
	r.Get("/admin/login", adminH.LoginPage)
	r.Post("/admin/login", adminH.LoginSubmit)
	r.Get("/admin/logout", adminH.Logout)

	// --- Admin routes (protected) ---
	r.Group(func(r chi.Router) {
		r.Use(adminH.AuthMiddleware)
		r.Get("/admin", adminH.Dashboard)
		r.Get("/admin/dashboard", adminH.Dashboard)
		r.Get("/admin/sections", adminH.SectionsOverview)
		r.Get("/admin/articles", adminH.ArticlesList)
		r.Get("/admin/articles/new", adminH.ArticleNew)
		r.Post("/admin/articles/new", adminH.ArticleCreate)
		r.Get("/admin/articles/edit/{id}", adminH.ArticleEdit)
		r.Post("/admin/articles/edit/{id}", adminH.ArticleUpdate)
		r.Post("/admin/articles/delete/{id}", adminH.ArticleDelete)
	})

	// --- Static files ---
	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static")))
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		if cfg.Env != "development" {
			w.Header().Set("Cache-Control", "public, max-age=86400")
		}
		fileServer.ServeHTTP(w, r)
	})

	// --- Health check ---
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	slog.Info("starting server", "port", cfg.Port, "env", cfg.Env, "db", cfg.DBType)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
