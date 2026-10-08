package handler

import (
	"context"
	"dr-tahoorian/internal/config"
	"dr-tahoorian/internal/domain"
	"dr-tahoorian/internal/middleware"
	"dr-tahoorian/internal/repository"
	"dr-tahoorian/internal/service"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	cfg      *config.Config
	repo     repository.Repository
	svc      *service.Service
	sessions *middleware.SessionStore
	tmplDir  string
}

func NewHandler(cfg *config.Config, repo repository.Repository, svc *service.Service, tmplDir string) *Handler {
	return &Handler{
		cfg:      cfg,
		repo:     repo,
		svc:      svc,
		sessions: middleware.NewSessionStore(),
		tmplDir:  tmplDir,
	}
}

type PageData struct {
	Title       string
	Description string
	Keywords    string
	OGImage     string
	OGType      string
	CurrentPath string
	Content     interface{}
	Data        interface{}
	CSRFToken   string
	Error       string
	Success     string
}

func (h *Handler) renderPage(w http.ResponseWriter, pageName string, data PageData) {
	tmplPath := filepath.Join(h.tmplDir, "pages", pageName+".html")
	basePath := filepath.Join(h.tmplDir, "layout", "base.html")
	headerPath := filepath.Join(h.tmplDir, "partials", "header.html")
	footerPath := filepath.Join(h.tmplDir, "partials", "footer.html")

	t, err := template.ParseFiles(basePath, tmplPath, headerPath, footerPath)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.ExecuteTemplate(w, "base", data)
}

func (h *Handler) renderAdmin(w http.ResponseWriter, r *http.Request, adminView string, data PageData) {
	sess := h.sessions.GetOrCreate(r, w)
	data.CSRFToken = sess.CSRFToken

	layoutPath := filepath.Join(h.tmplDir, "admin", "layout.html")
	viewPath := filepath.Join(h.tmplDir, "admin", adminView+".html")

	t, err := template.ParseFiles(layoutPath, viewPath)
	if err != nil {
		http.Error(w, "Admin template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.ExecuteTemplate(w, "admin_layout", data)
}

func (h *Handler) HandleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	about, stats, services, articles, gallery, _ := h.svc.GetHomeData(ctx)
	h.renderPage(w, "index", PageData{
		Title:       "دکتر حسین طهوریان | طب سنتی و رگ‌گیری، ثبت اختراع، سرمایه‌گذاری و مشاوره تخصصی",
		Description: "وب‌سایت رسمی دکتر حسین طهوریان — خدمات تخصصی طب سنتی و رگ‌گیری دستی، ۵ اختراع ثبت‌شده رسمی، سرمایه‌گذاری هلدینگ و مشاوره تخصصی مدیریت سازمانی.",
		CurrentPath: "/",
		Content: map[string]interface{}{
			"About":        about,
			"Stats":        stats,
			"Services":     services,
			"Articles":     articles,
			"Gallery":      gallery,
			"GalleryCount": len(gallery),
		},
	})
}

func (h *Handler) HandleAbout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	about, _ := h.svc.GetAboutData(ctx)
	h.renderPage(w, "about", PageData{
		Title:       "درباره دکتر حسین طهوریان | طب سنتی، مخترع و مشاور سرمایه‌گذاری هلدینگ",
		Description: "زندگینامه و دستاوردهای دکتر حسین طهوریان؛ تخصص در طب سنتی و فنون رگ‌گیری، ۵ اختراع ثبت‌شده ملی و بین‌المللی، مدال طلای مخترعان سوئیس و راهبری هلدینگ‌ها.",
		CurrentPath: "/about",
		Content: map[string]interface{}{
			"About": about,
		},
	})
}

func (h *Handler) HandleServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	services, packages, _ := h.svc.GetServicesData(ctx)
	h.renderPage(w, "services", PageData{
		Title:       "پکیج‌های جامع تخصصی | طب سنتی و رگ‌گیری، سرمایه‌گذاری، اختراع و مشاوره مدیریت",
		Description: "پکیج‌های جامع و اختصاصی دکتر حسین طهوریان: طب سنتی و رگ‌گیری تخصصی دستی، توسعه سرمایه‌گذاری و هلدینگ، تجاری‌سازی نوآوری و ثبت اختراع، و مشاوره تخصصی مدیریت.",
		CurrentPath: "/services",
		Content: map[string]interface{}{
			"Services": services,
			"Packages": packages,
		},
	})
}

func (h *Handler) HandleProjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	gallery, _ := h.svc.GetProjectsData(ctx)
	h.renderPage(w, "projects", PageData{
		Title:       "پروژه‌ها و هلدینگ‌های همکار | دکتر حسین طهوریان",
		Description: "هلدینگ سرآمد سرمایه ایلیا، گروه پتنت و پروژه‌های موفق سرمایه‌گذاری و دانش‌بنیان دکتر طهوریان.",
		CurrentPath: "/projects",
		Content: map[string]interface{}{
			"Projects": gallery,
		},
	})
}

func (h *Handler) HandleArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, categories, _ := h.svc.GetArticlesData(ctx)
	h.renderPage(w, "articles", PageData{
		Title:       "مقالات تخصصی و یادداشت‌ها | دکتر حسین طهوریان",
		Description: "مجموعه مقالات تخصصی دکتر حسین طهوریان در حوزه طب سنتی، نوآوری، ثبت اختراع، سرمایه‌گذاری خطرپذیر و مدیریت راهبردی کسب‌وکار.",
		CurrentPath: "/articles",
		Content: map[string]interface{}{
			"Articles":   articles,
			"Categories": categories,
		},
	})
}

func (h *Handler) HandleArticleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	article, related, err := h.svc.GetArticleDetail(ctx, slug)
	if err != nil || article == nil {
		http.Redirect(w, r, "/articles", http.StatusFound)
		return
	}

	h.renderPage(w, "article_detail", PageData{
		Title:       article.Title + " | دکتر حسین طهوریان",
		Description: article.Summary,
		CurrentPath: "/articles/" + slug,
		Content: map[string]interface{}{
			"Article": article,
			"Related": related,
		},
	})
}

func (h *Handler) HandleTeam(w http.ResponseWriter, r *http.Request) {
	h.renderPage(w, "team", PageData{
		Title:       "تیم و مشاوران همراه | دکتر حسین طهوریان",
		Description: "معرفی شبکه مشاوران، مدیران اجرایی و نخبگان علمی همکار در پروژه‌های سرمایه‌گذاری، ثبت اختراع و کلینیک تخصصی دکتر طهوریان.",
		CurrentPath: "/team",
	})
}

func (h *Handler) HandleContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contactInfo, _ := h.repo.GetContactInfo(ctx)
	h.renderPage(w, "contact", PageData{
		Title:       "تماس با دکتر حسین طهوریان | دریافت نوبت و مشاوره",
		Description: "راه‌های ارتباطی، نشانی دفتر مرکزی و فرم تماس جهت رزرو وقت معاینه طب سنتی، مشاوره سرمایه‌گذاری و راهبری هلدینگ.",
		CurrentPath: "/contact",
		Content: map[string]interface{}{
			"ContactInfo": contactInfo,
		},
	})
}

func (h *Handler) HandleContactSubmit(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	phone := strings.TrimSpace(r.FormValue("phone"))
	email := strings.TrimSpace(r.FormValue("email"))
	subject := strings.TrimSpace(r.FormValue("subject"))
	message := strings.TrimSpace(r.FormValue("message"))

	if name == "" || phone == "" {
		http.Error(w, "نام و شماره تماس الزامی است.", http.StatusBadRequest)
		return
	}

	contact := &domain.Contact{
		Name:      name,
		Phone:     phone,
		Email:     email,
		Subject:   subject,
		Message:   message,
		CreatedAt: time.Now(),
	}

	_ = h.svc.SubmitContact(r.Context(), contact)

	// Return htmx partial response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprint(w, `<div class="p-6 rounded-2xl bg-emerald-800 text-white text-center font-bold">پیام شما با موفقیت ثبت شد. در اسرع وقت با شما تماس گرفته خواهد شد.</div>`)
}

// Admin Handlers
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		sess, ok := h.sessions.Get(cookie.Value)
		if !ok || !sess.Authenticated {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		loginPath := filepath.Join(h.tmplDir, "admin", "login.html")
		t, err := template.ParseFiles(loginPath)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = t.Execute(w, PageData{Title: "ورود به پنل مدیریت"})
		return
	}

	_ = r.ParseForm()
	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))

	if username == h.cfg.AdminUser && password == h.cfg.AdminPass {
		sess := h.sessions.GetOrCreate(r, w)
		sess.Authenticated = true
		sess.Username = username
		http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
		return
	}

	loginPath := filepath.Join(h.tmplDir, "admin", "login.html")
	t, _ := template.ParseFiles(loginPath)
	_ = t.Execute(w, PageData{Title: "ورود به پنل مدیریت", Error: "نام کاربری یا رمز عبور اشتباه است."})
}

func (h *Handler) HandleAdminLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("admin_session")
	if err == nil {
		h.sessions.Delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "admin_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

func (h *Handler) HandleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, _ := h.repo.GetArticles(ctx, false)
	pubCount := 0
	for _, a := range articles {
		if a.IsPublished {
			pubCount++
		}
	}
	h.renderAdmin(w, r, "dashboard", PageData{
		Title:       "داشبورد مدیریت | طهوریان",
		CurrentPath: "/admin/dashboard",
		Data: map[string]interface{}{
			"TotalArticles":     len(articles),
			"PublishedArticles": pubCount,
			"DraftArticles":     len(articles) - pubCount,
			"RecentArticles":    articles,
		},
	})
}

func (h *Handler) HandleAdminArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, _ := h.repo.GetArticles(ctx, false)
	h.renderAdmin(w, r, "articles_list", PageData{
		Title:       "مدیریت مقالات | طهوریان",
		CurrentPath: "/admin/articles",
		Data: map[string]interface{}{
			"Articles": articles,
		},
	})
}

func (h *Handler) HandleAdminArticleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method == http.MethodGet {
		cats, _ := h.repo.GetCategories(ctx)
		h.renderAdmin(w, r, "article_form", PageData{
			Title:       "افزودن مقاله جدید | طهوریان",
			CurrentPath: "/admin/articles/new",
			Data: map[string]interface{}{
				"IsNew":      true,
				"Categories": cats,
			},
		})
		return
	}

	_ = r.ParseForm()
	title := strings.TrimSpace(r.FormValue("title"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	if slug == "" {
		slug = strings.ReplaceAll(strings.ToLower(title), " ", "-")
	}

	a := &domain.Article{
		Title:       title,
		Slug:        slug,
		Category:    r.FormValue("category"),
		Summary:     r.FormValue("summary"),
		Content:     r.FormValue("content"),
		ImageURL:    r.FormValue("image_url"),
		Author:      "دکتر حسین طهوریان",
		IsPublished: r.FormValue("is_published") == "true" || r.FormValue("is_published") == "on",
	}

	_ = h.repo.CreateArticle(ctx, a)
	http.Redirect(w, r, "/admin/articles", http.StatusFound)
}

func (h *Handler) HandleAdminArticleDelete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	_ = h.repo.DeleteArticle(r.Context(), id)
	http.Redirect(w, r, "/admin/articles", http.StatusFound)
}
