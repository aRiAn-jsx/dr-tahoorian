package main

import (
	"context"
	"dr-tahoorian/internal/config"
	"dr-tahoorian/internal/handler"
	"dr-tahoorian/internal/middleware"
	"dr-tahoorian/internal/repository/sqlite"
	"dr-tahoorian/internal/service"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()

	// SQLite Repository
	repo, err := sqlite.NewSQLiteRepo(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite repository: %v", err)
	}

	// Service Layer
	svc := service.NewService(repo, cfg.TelegramBotToken, cfg.TelegramChatID)

	// HTTP Handlers
	tmplDir := "views"
	h := handler.NewHandler(cfg, repo, svc, tmplDir)

	// Router
	r := chi.NewRouter()

	// Middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.SecurityHeaders)

	// Static Assets
	staticDir := http.Dir("web/static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticDir)))

	// Public Routes
	r.Get("/", h.HandleHome)
	r.Get("/about", h.HandleAbout)
	r.Get("/services", h.HandleServices)
	r.Get("/projects", h.HandleProjects)
	r.Get("/articles", h.HandleArticles)
	r.Get("/articles/{slug}", h.HandleArticleDetail)
	r.Get("/team", h.HandleTeam)
	r.Get("/contact", h.HandleContact)
	r.Post("/contact", h.HandleContactSubmit)

	// Admin Auth Routes
	r.Get("/admin/login", h.HandleAdminLogin)
	r.Post("/admin/login", h.HandleAdminLogin)
	r.Get("/admin/logout", h.HandleAdminLogout)

	// Admin Protected Group
	r.Group(func(admin chi.Router) {
		admin.Use(h.RequireAuth)
		admin.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
		})
		admin.Get("/admin/dashboard", h.HandleAdminDashboard)
		admin.Get("/admin/articles", h.HandleAdminArticles)
		admin.Get("/admin/articles/new", h.HandleAdminArticleCreate)
		admin.Post("/admin/articles/new", h.HandleAdminArticleCreate)
		admin.Post("/admin/articles/delete/{id}", h.HandleAdminArticleDelete)
	})

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Dr. Tahoorian Go server running on http://%s in %s mode", addr, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen failed: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Go server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Go server gracefully stopped.")
}
