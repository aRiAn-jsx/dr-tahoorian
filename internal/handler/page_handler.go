package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tahoorian/tahoorian/internal/domain"
	"github.com/tahoorian/tahoorian/internal/service"
)

type Handler struct {
	svc *service.Service
}

type PageData struct {
	Title        string
	Content      interface{}
	CurrentPath  string
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, name string, data PageData) {
	data.CurrentPath = r.URL.Path
	baseDir, err := os.Getwd()
	if err != nil {
		http.Error(w, "Cannot determine working directory", http.StatusInternalServerError)
		return
	}

	layoutPath := filepath.Join(baseDir, "web", "templates", "layout", "base.html")
	headerPath := filepath.Join(baseDir, "web", "templates", "partials", "header.html")
	footerPath := filepath.Join(baseDir, "web", "templates", "partials", "footer.html")
	pagePath := filepath.Join(baseDir, "web", "templates", "pages", name+".html")

	tmpl, err := template.New(name).Funcs(template.FuncMap{
		"add1": func(value int) int { return value + 1 },
	}).ParseFiles(layoutPath, headerPath, footerPath, pagePath)
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
	gallery, _ := h.svc.ListGallery("")
	h.render(w, r, "index", PageData{
		Title: "خانه | طهوریان",
		Content: map[string]interface{}{
			"Stats":    stats,
			"Services": services,
			"Gallery":  gallery,
		},
	})
}

func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	about, _ := h.svc.GetAbout()
	gallery, _ := h.svc.ListGallery("")
	h.render(w, r, "about", PageData{
		Title: "درباره ما | طهوریان",
		Content: map[string]interface{}{
			"About":   about,
			"Gallery": gallery,
		},
	})
}

func (h *Handler) Services(w http.ResponseWriter, r *http.Request) {
	services, _ := h.svc.ListServices()
	h.render(w, r, "services", PageData{
		Title:   "خدمات | طهوریان",
		Content: services,
	})
}

func (h *Handler) Projects(w http.ResponseWriter, r *http.Request) {
	gallery, _ := h.svc.ListGallery("")
	h.render(w, r, "projects", PageData{
		Title: "پروژه‌ها | طهوریان",
		Content: map[string]interface{}{"Gallery": gallery},
	})
}

func (h *Handler) Team(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "team", PageData{
		Title: "تیم من | طهوریان",
	})
}

func (h *Handler) Contact(w http.ResponseWriter, r *http.Request) {
	info, _ := h.svc.GetContactInfo()
	h.render(w, r, "contact", PageData{
		Title:   "تماس با ما | طهوریان",
		Content: info,
	})
}

func (h *Handler) SubmitContact(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "پیام شما با موفقیت ارسال شد"})
}
