package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/tahoorian/tahoorian/internal/handler"
	"github.com/tahoorian/tahoorian/internal/middleware"
	memrepo "github.com/tahoorian/tahoorian/internal/repository/memory"
	"github.com/tahoorian/tahoorian/internal/service"
)

func setupRouter() (*chi.Mux, *service.Service) {
	repo := memrepo.New()
	svc := service.New(repo)
	h := handler.New(svc, true) // devMode = true for testing
	sm := middleware.NewSessionManager("test-secret-key-123")
	adminH := handler.NewAdmin(svc, sm, true)

	r := chi.NewRouter()

	// Public routes
	r.Get("/", h.Index)
	r.Get("/about", h.About)
	r.Get("/services", h.Services)
	r.Get("/projects", h.Projects)
	r.Get("/articles", h.Articles)
	r.Get("/articles/{slug}", h.ArticleDetail)
	r.Get("/team", h.Team)
	r.Get("/contact", h.Contact)
	r.Post("/api/contact", h.SubmitContact)

	// Admin routes
	r.Get("/admin/login", adminH.LoginPage)
	r.Post("/admin/login", adminH.LoginSubmit)
	r.Get("/admin/logout", adminH.Logout)

	r.Group(func(r chi.Router) {
		r.Use(adminH.AuthMiddleware)
		r.Get("/admin", adminH.Dashboard)
		r.Get("/admin/dashboard", adminH.Dashboard)
		r.Get("/admin/articles", adminH.ArticlesList)
	})

	return r, svc
}

func TestPublicRoutes(t *testing.T) {
	router, _ := setupRouter()

	endpoints := []string{
		"/",
		"/about",
		"/services",
		"/projects",
		"/articles",
		"/articles/principles-of-strategic-holding-management",
		"/team",
		"/contact",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest("GET", ep, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 OK for %s, got %d", ep, rec.Code)
		}
	}
}

func TestSubmitContactAPI(t *testing.T) {
	router, _ := setupRouter()

	form := url.Values{}
	form.Add("name", "کاربر تست")
	form.Add("email", "test@example.com")
	form.Add("message", "پیام جدید تست")

	req := httptest.NewRequest("POST", "/api/contact", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for contact submission, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "success") {
		t.Errorf("expected success response, got %s", rec.Body.String())
	}
}

func TestAdminAuthFlow(t *testing.T) {
	router, _ := setupRouter()

	// Unauthenticated request to protected route
	req := httptest.NewRequest("GET", "/admin/dashboard", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected 303 redirect for unauthed access, got %d", rec.Code)
	}

	// Login page GET
	req = httptest.NewRequest("GET", "/admin/login", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin login page, got %d", rec.Code)
	}

	// Login submit with invalid credentials
	loginForm := url.Values{}
	loginForm.Add("username", "wrong")
	loginForm.Add("password", "wrong")
	req = httptest.NewRequest("POST", "/admin/login", strings.NewReader(loginForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "نادرست") {
		t.Errorf("expected failure message on invalid login")
	}

	// Login submit with correct credentials
	loginForm.Set("username", "admin")
	loginForm.Set("password", "admin123")
	req = httptest.NewRequest("POST", "/admin/login", strings.NewReader(loginForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect after successful login, got %d", rec.Code)
	}

	// Extract cookie
	cookies := rec.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "admin_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("expected admin_session cookie to be set")
	}

	// Request protected route with valid cookie
	req = httptest.NewRequest("GET", "/admin/dashboard", nil)
	req.AddCookie(sessionCookie)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK for authenticated dashboard request, got %d", rec.Code)
	}
}
