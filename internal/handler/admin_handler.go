package handler

import (
	"crypto/subtle"
	"fmt"
	"html/template"
	"log/slog"
	"net"
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
	svc          *service.Service
	sm           *middleware.SessionManager
	tmplCache    map[string]*template.Template
	tmplMu       sync.RWMutex
	baseDir      string
	devMode      bool
	loginMu      sync.Mutex
	loginAttempts map[string]*loginAttempt
}

type loginAttempt struct {
	count    int
	lockedAt time.Time
}

// Rate limiting for admin login attempts
const (
	loginMaxAttempts = 5
	loginLockout     = 15 * time.Minute
)

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
		svc:           svc,
		sm:            sm,
		tmplCache:     make(map[string]*template.Template),
		baseDir:       baseDir,
		devMode:       devMode,
		loginAttempts: make(map[string]*loginAttempt),
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

	loginPath := filepath.Join(findRootDir(), "web", "templates", "admin", "login.html")
	tmpl, err := template.ParseFiles(loginPath)
	if err != nil {
		http.Error(w, "Template error: "+err.Error()+" ("+loginPath+")", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, AdminPageData{Title: "ورود به پنل مدیریت"})
}

func (a *AdminHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	// --- Brute-force protection (per-IP) ---
	clientIP := r.RemoteAddr
	if host, _, err := net.SplitHostPort(clientIP); err == nil {
		clientIP = host
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	if a.loginLocked(clientIP) {
		http.Error(w, "تعداد تلاش‌ها زیاد شد؛ لطفاً ۱۵ دقیقه بعد دوباره تلاش کنید", http.StatusTooManyRequests)
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
		slog.Warn("ADMIN_PASS not set — using weak default 'admin'. Set ADMIN_PASS env var in production!")
	}

	// Constant-time comparison to prevent timing attacks
	userMatch := subtle.ConstantTimeCompare([]byte(username), []byte(expectedUser)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(password), []byte(expectedPass)) == 1

	if userMatch && passMatch {
		// Reset attempts on success
		a.recordLoginAttempt(clientIP, true)

		session := a.sm.Create()
		session.Data["authenticated"] = true
		session.Data["username"] = username
		// Generate a CSRF token for this session
		session.Data["csrf_token"] = session.ID[:16]

		cookie := &http.Cookie{
			Name:     "admin_session",
			Value:    session.ID,
			Path:     "/",
			HttpOnly: true,
			Secure:  true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86400 * 7,
		}
		// In development (no TLS) Secure would block the cookie; only set Secure if TLS is in use.
		if strings.ToLower(os.Getenv("ENV")) != "production" && strings.ToLower(os.Getenv("ENV")) != "prod" {
			cookie.Secure = false
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
		return
	}

	// Failed attempt — record and add artificial delay
	a.recordLoginAttempt(clientIP, false)
	time.Sleep(500 * time.Millisecond) // mitigate user enumeration via timing

	loginPath := filepath.Join(findRootDir(), "web", "templates", "admin", "login.html")
	tmpl, err := template.ParseFiles(loginPath)
	if err != nil {
		http.Error(w, "Template error: "+err.Error()+" ("+loginPath+")", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, AdminPageData{
		Title: "ورود به پنل مدیریت",
		Error: "نام کاربری یا رمز عبور اشتباه است",
	})
}

func (a *AdminHandler) loginLocked(ip string) bool {
	a.loginMu.Lock()
	defer a.loginMu.Unlock()
	at, ok := a.loginAttempts[ip]
	if !ok {
		return false
	}
	if time.Since(at.lockedAt) > loginLockout {
		delete(a.loginAttempts, ip)
		return false
	}
	return at.count >= loginMaxAttempts
}

func (a *AdminHandler) recordLoginAttempt(ip string, success bool) {
	a.loginMu.Lock()
	defer a.loginMu.Unlock()
	if success {
		delete(a.loginAttempts, ip)
		return
	}
	at := a.loginAttempts[ip]
	if at == nil {
		at = &loginAttempt{}
		a.loginAttempts[ip] = at
	}
	at.count++
	if at.count >= loginMaxAttempts {
		at.lockedAt = time.Now()
	}
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
	cats, _ := a.svc.ListCategories()
	a.render(w, r, "article_form", AdminPageData{
		Title:     "افزودن مقاله جدید | طهوریان",
		CSRFToken: a.csrfToken(r),
		Data: map[string]interface{}{
			"IsNew":      true,
			"Article":    domain.Article{Author: "دکتر حسین طهوریان", IsPublished: true},
			"Categories": cats,
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
	category := strings.TrimSpace(r.FormValue("category"))
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
					Category:    category,
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
		Category:    category,
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

	cats, _ := a.svc.ListCategories()
	a.render(w, r, "article_form", AdminPageData{
		Title:     fmt.Sprintf("ویرایش مقاله: %s", article.Title),
		CSRFToken: a.csrfToken(r),
		Data: map[string]interface{}{
			"IsNew":      false,
			"Article":    *article,
			"Categories": cats,
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
	category := strings.TrimSpace(r.FormValue("category"))
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
		Category:    category,
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

// CategoriesList renders the category management page. It reads the success/failure
// message from the querystring (set by redirect after create/delete).
func (a *AdminHandler) CategoriesList(w http.ResponseWriter, r *http.Request) {
	cats, _ := a.svc.ListCategories()
	successMsg := ""
	switch r.URL.Query().Get("msg") {
	case "created":
		successMsg = "دسته‌بندی جدید با موفقیت ایجاد شد."
	case "deleted":
		successMsg = "دسته‌بندی حذف شد."
	}
	a.render(w, r, "categories", AdminPageData{
		Title:     "دسته‌بندی مقالات | طهوریان",
		Success:   successMsg,
		CSRFToken: a.csrfToken(r),
		Data: map[string]interface{}{
			"Categories": cats,
		},
	})
}

// CategoryCreate adds a new article category. The slug is validated/normalized in
// the service layer so a human-readable title can be turned into a url-safe slug.
func (a *AdminHandler) CategoryCreate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	if !a.validateCSRF(r) {
		http.Error(w, "درخواست نامعتبر (CSRF)", http.StatusForbidden)
		return
	}

	cat := &domain.ArticleCategory{
		Slug:  strings.TrimSpace(r.FormValue("slug")),
		Title: strings.TrimSpace(r.FormValue("title")),
	}
	if err := a.svc.CreateCategory(cat); err != nil {
		cats, _ := a.svc.ListCategories()
		a.render(w, r, "categories", AdminPageData{
			Title:     "دسته‌بندی مقالات | طهوریان",
			Error:     "خطا در ایجاد دسته‌بندی: " + err.Error(),
			CSRFToken: a.csrfToken(r),
			Data: map[string]interface{}{
				"Categories": cats,
				"FormTitle":  cat.Title,
				"FormSlug":   cat.Slug,
			},
		})
		return
	}
	http.Redirect(w, r, "/admin/categories?msg=created", http.StatusSeeOther)
}

// CategoryDelete removes a category. Deleting a category does not remove articles
// that referenced it; their card simply falls back to displaying the slug.
func (a *AdminHandler) CategoryDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin/categories", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err == nil {
		if !a.validateCSRF(r) {
			http.Error(w, "درخواست نامعتبر (CSRF)", http.StatusForbidden)
			return
		}
	}
	_ = a.svc.DeleteCategory(id)
	http.Redirect(w, r, "/admin/categories?msg=deleted", http.StatusSeeOther)
}
