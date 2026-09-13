package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tahoorian/tahoorian/internal/domain"
	"github.com/tahoorian/tahoorian/internal/service"
)

type Handler struct {
	svc       *service.Service
	tmplCache map[string]*template.Template
	tmplMu    sync.RWMutex
	baseDir   string
	devMode   bool
}

type PageData struct {
	Title       string
	Description string
	CurrentPath string
	Content     any
}

var pageFuncs = template.FuncMap{
	"hasPrefix": strings.HasPrefix,
	"add1":      func(value int) int { return value + 1 },
	"formatDate": func(t time.Time) string {
		if t.IsZero() {
			return ""
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
}

func findRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "web", "templates")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "."
}

func New(svc *service.Service, devMode bool) *Handler {
	baseDir := findRootDir()
	h := &Handler{
		svc:       svc,
		tmplCache: make(map[string]*template.Template),
		baseDir:   baseDir,
		devMode:   devMode,
	}
	pages := []string{"index", "about", "services", "articles", "article_detail", "team", "contact"}
	for _, p := range pages {
		if _, err := h.loadTemplate(p); err != nil {
			_ = err
		}
	}
	return h
}

func (h *Handler) loadTemplate(name string) (*template.Template, error) {
	if !h.devMode {
		h.tmplMu.RLock()
		if t, ok := h.tmplCache[name]; ok {
			h.tmplMu.RUnlock()
			return t, nil
		}
		h.tmplMu.RUnlock()
	}

	layoutPath := filepath.Join(h.baseDir, "web", "templates", "layout", "base.html")
	headerPath := filepath.Join(h.baseDir, "web", "templates", "partials", "header.html")
	footerPath := filepath.Join(h.baseDir, "web", "templates", "partials", "footer.html")
	pagePath := filepath.Join(h.baseDir, "web", "templates", "pages", name+".html")

	tmpl, err := template.New("base").Funcs(pageFuncs).ParseFiles(layoutPath, headerPath, footerPath, pagePath)
	if err != nil {
		return nil, err
	}

	if !h.devMode {
		h.tmplMu.Lock()
		h.tmplCache[name] = tmpl
		h.tmplMu.Unlock()
	}
	return tmpl, nil
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, name string, data PageData) {
	data.CurrentPath = r.URL.Path
	tmpl, err := h.loadTemplate(name)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	stats, _ := h.svc.ListStats()
	services, _ := h.svc.ListServices()
	articles, _ := h.svc.ListPublishedArticles()
	h.render(w, r, "index", PageData{
		Title:       "خانه | طهوریان",
		Description: "هلدینگ دکتر حسین طهوریان — ارائه خدمات تخصصی و پروژه‌های موفق",
		Content: map[string]interface{}{
			"Stats":    stats,
			"Services": services,
			"Articles": articles,
		},
	})
}

func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	about, _ := h.svc.GetAbout()
	h.render(w, r, "about", PageData{
		Title:       "درباره ما | طهوریان",
		Description: "آشنایی با دکتر حسین طهوریان و هلدینگ‌های مرتبط",
		Content: map[string]interface{}{
			"About": about,
		},
	})
}

func (h *Handler) Services(w http.ResponseWriter, r *http.Request) {
	services, _ := h.svc.ListServices()
	h.render(w, r, "services", PageData{
		Title:       "خدمات | طهوریان",
		Description: "خدمات تخصصی هلدینگ طهوریان",
		Content:     services,
	})
}

func (h *Handler) Projects(w http.ResponseWriter, r *http.Request) {
	projects, _ := h.svc.ListGallery("")
	h.render(w, r, "projects", PageData{
		Title:       "پروژه‌ها | طهوریان",
		Description: "نمونه پروژه‌های انجام شده توسط هلدینگ طهوریان",
		Content: map[string]interface{}{
			"Projects": projects,
		},
	})
}

func (h *Handler) Articles(w http.ResponseWriter, r *http.Request) {
	articles, _ := h.svc.ListPublishedArticles()
	h.render(w, r, "articles", PageData{
		Title:       "مقالات | طهوریان",
		Description: "مقالات و مطالب تخصصی دکتر حسین طهوریان",
		Content: map[string]interface{}{
			"Articles": articles,
		},
	})
}

func (h *Handler) ArticleDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	article, err := h.svc.GetArticleBySlug(slug)
	if err != nil || article == nil {
		http.Redirect(w, r, "/articles", http.StatusSeeOther)
		return
	}
	all, _ := h.svc.ListPublishedArticles()
	var recent []domain.Article
	for _, a := range all {
		if a.Slug != slug {
			recent = append(recent, a)
			if len(recent) >= 4 {
				break
			}
		}
	}
	h.render(w, r, "article_detail", PageData{
		Title:       article.Title + " | طهوریان",
		Description: article.Excerpt,
		Content: map[string]interface{}{
			"Article":        article,
			"RecentArticles": recent,
		},
	})
}

func (h *Handler) Team(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "team", PageData{
		Title:       "تیم ما | طهوریان",
		Description: "تیم متخصص هلدینگ طهوریان",
	})
}

func (h *Handler) Contact(w http.ResponseWriter, r *http.Request) {
	info, _ := h.svc.GetContactInfo()
	h.render(w, r, "contact", PageData{
		Title:       "تماس با ما | دکتر حسین طهوریان",
		Description: "راه‌های ارتباطی با هلدینگ طهوریان",
		CurrentPath: "/contact",
		Content:     info,
	})
}

func (h *Handler) SubmitContact(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`<div class="contact-response error"><i data-lucide="alert-circle"></i><div><strong>خطا در دریافت اطلاعات</strong><p>لطفاً فیلدهای ضروری را به درستی تکمیل فرمایید.</p></div></div>`))
			return
		}
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	contact := &domain.Contact{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Phone:   r.FormValue("phone"),
		Subject: r.FormValue("subject"),
		Message: r.FormValue("message"),
	}
	if err := h.svc.SubmitContact(contact); err != nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`<div class="contact-response error"><i data-lucide="alert-circle"></i><div><strong>خطا در ثبت پیام</strong><p>` + htmlEscape(err.Error()) + `</p></div></div>`))
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<div class="contact-response success"><i data-lucide="check-circle-2"></i><div><strong>پیام شما با موفقیت ثبت شد</strong><p>درخواست شما دریافت شد و در کمتر از ۲۴ ساعت کاری با شما تماس خواهیم گرفت.</p></div></div>`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "پیام شما با موفقیت ارسال شد"})
}

func htmlEscape(s string) string {
	return template.HTMLEscapeString(s)
}