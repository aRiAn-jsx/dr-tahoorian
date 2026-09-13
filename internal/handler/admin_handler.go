package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tahoorian/tahoorian/internal/domain"
	"github.com/tahoorian/tahoorian/internal/middleware"
	"github.com/tahoorian/tahoorian/internal/service"
)

type AdminHandler struct {
	svc       *service.Service
	sm        *middleware.SessionManager
	tmplCache map[string]*template.Template
	tmplMu    sync.RWMutex
	baseDir   string
	devMode   bool
}

type AdminPageData struct {
	Title       string
	CurrentPath string
	Error       string
	Success     string
	Data        interface{}
	CSRFToken   string
}

var adminFuncs = template.FuncMap{
	"formatDate": func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		return t.Format("2006/01/02 15:04")
	},
	"formatDateShort": func(t time.Time) string {
		if t.IsZero() {
			return "-"
		}
		return t.Format("2006/01/02")
	},
	"truncate": func(s string, length int) string {
		runes := []rune(s)
		if len(runes) <= length {
			return s
		}
		return string(runes[:length]) + "..."
	},
	"add1": func(i int) int {
		return i + 1
	},
}

func NewAdmin(svc *service.Service, sm *middleware.SessionManager, devMode bool) *AdminHandler {
	baseDir := findRootDir()
	return &AdminHandler{
		svc:       svc,
		sm:        sm,
		tmplCache: make(map[string]*template.Template),
		baseDir:   baseDir,
		devMode:   devMode,
	}
}

func (a *AdminHandler) loadTemplate(pageName string) (*template.Template, error) {
	if !a.devMode {
		a.tmplMu.RLock()
		if t, ok := a.tmplCache[pageName]; ok {
			a.tmplMu.RUnlock()
			return t, nil
		}
		a.tmplMu.RUnlock()
	}

	layoutPath := filepath.Join(a.baseDir, "web", "templates", "admin", "layout.html")
	pagePath := filepath.Join(a.baseDir, "web", "templates", "admin", pageName+".html")

	tmpl, err := template.New("admin_layout").Funcs(adminFuncs).ParseFiles(layoutPath, pagePath)
	if err != nil {
		return nil, err
	}

	if !a.devMode {
		a.tmplMu.Lock()
		a.tmplCache[pageName] = tmpl
		a.tmplMu.Unlock()
	}
	return tmpl, nil
}

func (a *AdminHandler) render(w http.ResponseWriter, r *http.Request, pageName string, data AdminPageData) {
	data.CurrentPath = r.URL.Path
	tmpl, err := a.loadTemplate(pageName)
	if err != nil {
		http.Error(w, "Admin Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "admin_layout", data); err != nil {
		http.Error(w, "Admin Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

// AuthMiddleware uses the SessionManager for proper session-based authentication.
func (a *AdminHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_session")
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		session, ok := a.sm.Get(cookie.Value)
		if !ok {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		authed, _ := session.Data["authenticated"].(bool)
		if !authed {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// csrfToken returns the CSRF token stored in the admin session, or empty string if not found.
func (a *AdminHandler) csrfToken(r *http.Request) string {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		return ""
	}
	session, ok := a.sm.Get(cookie.Value)
	if !ok {
		return ""
	}
	tok, _ := session.Data["csrf_token"].(string)
	return tok
}

// validateCSRF checks the CSRF token on mutating requests.
func (a *AdminHandler) validateCSRF(r *http.Request) bool {
	expected := a.csrfToken(r)
	if expected == "" {
		return false
	}
	submitted := r.FormValue("_csrf")
	return submitted == expected
}

func (a *AdminHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Already logged in? redirect to dashboard.
	cookie, err := r.Cookie("admin_session")
	if err == nil {
		if session, ok := a.sm.Get(cookie.Value); ok {
			if authed, _ := session.Data["authenticated"].(bool); authed {
				http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
				return
			}
		}
	}

	baseDir, _ := os.Getwd()
	loginPath := filepath.Join(baseDir, "web", "templates", "admin", "login.html")
	tmpl, err := template.ParseFiles(loginPath)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, AdminPageData{Title: "ورود به پنل مدیریت"})
}

func (a *AdminHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	expectedUser := os.Getenv("ADMIN_USER")
	if expectedUser == "" {
		expectedUser = "admin"
	}
	expectedPass := os.Getenv("ADMIN_PASS")
	if expectedPass == "" {
		expectedPass = "admin"
	}

	if username == expectedUser && password == expectedPass {
		session := a.sm.Create()
		session.Data["authenticated"] = true
		session.Data["username"] = username
		// Generate a CSRF token for this session
		session.Data["csrf_token"] = session.ID[:16]

		http.SetCookie(w, &http.Cookie{
			Name:     "admin_session",
			Value:    session.ID,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400 * 7,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	baseDir, _ := os.Getwd()
	loginPath := filepath.Join(baseDir, "web", "templates", "admin", "login.html")
	tmpl, _ := template.ParseFiles(loginPath)
	tmpl.Execute(w, AdminPageData{
		Title: "ورود به پنل مدیریت",
		Error: "نام کاربری یا رمز عبور اشتباه است",
	})
}

func (a *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("admin_session")
	if err == nil {
		a.sm.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "admin_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

func (a *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	articles, _ := a.svc.ListArticles()
	publishedCount := 0
	for _, art := range articles {
		if art.IsPublished {
			publishedCount++
		}
	}

	recentArticles := articles
	if len(recentArticles) > 5 {
		recentArticles = recentArticles[:5]
	}

	a.render(w, r, "dashboard", AdminPageData{
		Title:     "داشبورد مدیریت | طهوریان",
		CSRFToken: a.csrfToken(r),
		Data: map[string]interface{}{
			"TotalArticles":     len(articles),
			"PublishedArticles": publishedCount,
			"DraftArticles":     len(articles) - publishedCount,
			"RecentArticles":    recentArticles,
			"ActiveSections":    1,
			"TotalSections":     5,
		},
	})
}

func (a *AdminHandler) SectionsOverview(w http.ResponseWriter, r *http.Request) {
	articles, _ := a.svc.ListArticles()
	a.render(w, r, "sections", AdminPageData{
		Title:     "مدیریت بخش‌های سایت | طهوریان",
		CSRFToken: a.csrfToken(r),
		Data: map[string]interface{}{
			"ArticleCount": len(articles),
		},
	})
}

func (a *AdminHandler) ArticlesList(w http.ResponseWriter, r *http.Request) {
	articles, err := a.svc.ListArticles()
	if err != nil {
		articles = []domain.Article{}
	}

	msg := r.URL.Query().Get("msg")
	successMsg := ""
	switch msg {
	case "created":
		successMsg = "مقاله جدید با موفقیت ایجاد شد."
	case "updated":
		successMsg = "مقاله با موفقیت ویرایش شد."
	case "deleted":
		successMsg = "مقاله با موفقیت حذف شد."
	}

	a.render(w, r, "articles_list", AdminPageData{
		Title:     "مدیریت مقالات | طهوریان",
		Success:   successMsg,
		CSRFToken: a.csrfToken(r),
		Data: map[string]interface{}{
			"Articles": articles,
		},
	})
}

func (a *AdminHandler) ArticleNew(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "article_form", AdminPageData{
		Title:     "افزودن مقاله جدید | طهوریان",
		CSRFToken: a.csrfToken(r),
		Data: map[string]interface{}{
			"IsNew":   true,
			"Article": domain.Article{Author: "دکتر حسین طهوریان", IsPublished: true},
		},
	})
}

func (a *AdminHandler) ArticleCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	if !a.validateCSRF(r) {
		http.Error(w, "درخواست نامعتبر (CSRF)", http.StatusForbidden)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	summary := strings.TrimSpace(r.FormValue("summary"))
	content := strings.TrimSpace(r.FormValue("content"))
	imageURL := strings.TrimSpace(r.FormValue("image_url"))
	author := strings.TrimSpace(r.FormValue("author"))
	if author == "" {
		author = "دکتر حسین طهوریان"
	}
	isPublished := r.FormValue("is_published") == "1" || r.FormValue("is_published") == "true" || r.FormValue("is_published") == "on"

	if title == "" {
		a.render(w, r, "article_form", AdminPageData{
			Title:     "افزودن مقاله جدید | طهوریان",
			Error:     "عنوان مقاله نمی‌تواند خالی باشد.",
			CSRFToken: a.csrfToken(r),
			Data: map[string]interface{}{
				"IsNew": true,
				"Article": domain.Article{
					Title:       title,
					Slug:        slug,
					Summary:     summary,
					Content:     content,
					ImageURL:    imageURL,
					Author:      author,
					IsPublished: isPublished,
				},
			},
		})
		return
	}

	if slug == "" {
		slug = strings.ToLower(strings.ReplaceAll(title, " ", "-"))
	}

	article := &domain.Article{
		Title:       title,
		Slug:        slug,
		Summary:     summary,
		Content:     content,
		ImageURL:    imageURL,
		Author:      author,
		IsPublished: isPublished,
	}

	if err := a.svc.CreateArticle(article); err != nil {
		a.render(w, r, "article_form", AdminPageData{
			Title:     "افزودن مقاله جدید | طهوریان",
			Error:     "خطا در ذخیره مقاله: " + err.Error(),
			CSRFToken: a.csrfToken(r),
			Data: map[string]interface{}{
				"IsNew":   true,
				"Article": *article,
			},
		})
		return
	}

	http.Redirect(w, r, "/admin/articles?msg=created", http.StatusSeeOther)
}

func (a *AdminHandler) ArticleEdit(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin/articles", http.StatusSeeOther)
		return
	}

	article, err := a.svc.GetArticleByID(id)
	if err != nil || article == nil {
		http.Redirect(w, r, "/admin/articles", http.StatusSeeOther)
		return
	}

	a.render(w, r, "article_form", AdminPageData{
		Title:     fmt.Sprintf("ویرایش مقاله: %s", article.Title),
		CSRFToken: a.csrfToken(r),
		Data: map[string]interface{}{
			"IsNew":   false,
			"Article": *article,
		},
	})
}

func (a *AdminHandler) ArticleUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin/articles", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	if !a.validateCSRF(r) {
		http.Error(w, "درخواست نامعتبر (CSRF)", http.StatusForbidden)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	summary := strings.TrimSpace(r.FormValue("summary"))
	content := strings.TrimSpace(r.FormValue("content"))
	imageURL := strings.TrimSpace(r.FormValue("image_url"))
	author := strings.TrimSpace(r.FormValue("author"))
	if author == "" {
		author = "دکتر حسین طهوریان"
	}
	isPublished := r.FormValue("is_published") == "1" || r.FormValue("is_published") == "true" || r.FormValue("is_published") == "on"

	article := &domain.Article{
		ID:          id,
		Title:       title,
		Slug:        slug,
		Summary:     summary,
		Content:     content,
		ImageURL:    imageURL,
		Author:      author,
		IsPublished: isPublished,
	}

	if err := a.svc.UpdateArticle(article); err != nil {
		a.render(w, r, "article_form", AdminPageData{
			Title:     "ویرایش مقاله",
			Error:     "خطا در ویرایش مقاله: " + err.Error(),
			CSRFToken: a.csrfToken(r),
			Data: map[string]interface{}{
				"IsNew":   false,
				"Article": *article,
			},
		})
		return
	}

	http.Redirect(w, r, "/admin/articles?msg=updated", http.StatusSeeOther)
}

func (a *AdminHandler) ArticleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin/articles", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err == nil {
		if !a.validateCSRF(r) {
			http.Error(w, "درخواست نامعتبر (CSRF)", http.StatusForbidden)
			return
		}
	}

	_ = a.svc.DeleteArticle(id)
	http.Redirect(w, r, "/admin/articles?msg=deleted", http.StatusSeeOther)
}
