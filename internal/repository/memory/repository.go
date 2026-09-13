package memory

import (
	"sync"
	"time"

	"github.com/tahoorian/tahoorian/internal/domain"
)

type Repository struct {
	mu            sync.RWMutex
	about         *domain.About
	services      []domain.Service
	contacts      []domain.Contact
	gallery       []domain.GalleryItem
	stats         []domain.Stat
	contactInfo   *domain.ContactInfo
	articles      []domain.Article
	nextArticleID int
}

func New() *Repository {
	now := time.Now()
	return &Repository{
		nextArticleID: 4,
		articles: []domain.Article{
			{
				ID:          1,
				Title:       "اصول بنیادین در مدیریت استراتژیک هلدینگ‌ها",
				Slug:        "principles-of-strategic-holding-management",
				Summary:     "چگونه شرکت‌های مادر و هلدینگ‌ها می‌توانند هم‌افزایی ارزش را بین شرکت‌های تابعه ایجاد و هدایت کنند.",
				Content:     "مدیریت استراتژیک در هلدینگ‌ها نیازمند تفکیک دقیق میان سطح استراتژی کسب‌وکار و سطح استراتژی شرکتی است. در یک گروه اقتصادی موفق، هدایت سرمایه‌ها بر مبنای تحلیل دقیق جریان‌های نقدی، هم‌افزایی میان‌رشته‌ای، و توانمندسازی مدیران ارشد شرکت‌های تابعه انجام می‌پذیرد...",
				ImageURL:    "/static/images/کتاب-های-سایت-دکتر-طهوریان-1024x569.jpg",
				Author:      "دکتر حسین طهوریان",
				IsPublished: true,
				CreatedAt:   now.AddDate(0, -1, -5),
				UpdatedAt:   now.AddDate(0, -1, -5),
			},
			{
				ID:          2,
				Title:       "تبدیل نوآوری به ثبت اختراع و محصول تجاری",
				Slug:        "turning-innovation-into-patents-and-products",
				Summary:     "مسیر تجاری‌سازی ایده‌ها و اختراعات از فرضیه تا تولید صنعتی و ورود به بازار رقابتی.",
				Content:     "ایده‌ها تا زمانی که به یک مدل قابل اتکا و تکرارپذیر تبدیل نشوند، صرفاً در حد پتانسیل باقی می‌مانند. تجربه ثبت بیش از ۵ اختراع و پیاده‌سازی صنعتی نشان می‌دهد که فرآیند تحقیق و توسعه باید با نیازمندی‌های بازار و تحلیل زنجیره ارزش پیوند نزدیک داشته باشد...",
				ImageURL:    "/static/images/اختراعات-دکتر-سایت.png",
				Author:      "دکتر حسین طهوریان",
				IsPublished: true,
				CreatedAt:   now.AddDate(0, 0, -12),
				UpdatedAt:   now.AddDate(0, 0, -12),
			},
			{
				ID:          3,
				Title:       "رهبری تیم‌های ارزش‌آفرین در شرایط ابهام",
				Slug:        "leadership-in-uncertainty",
				Summary:     "ابزارهای کلیدی یک مدیر در تصمیم‌گیری‌های حساس و هدایت انگیزه سازمان در شرایط نااطمینانی اقتصادی.",
				Content:     "رهبری در شرایط ابهام صرفاً پیش‌بینی دقیق آینده نیست؛ بلکه خلق قابلیت انطباق‌پذیری بالا در سازمان است. تیم‌هایی که شفافیت هدف و استقلال در تصمیم‌گیری دارند، بحران‌ها را به سکوی پرتاب تبدیل می‌کنند...",
				ImageURL:    "/static/images/شرکت.png",
				Author:      "دکتر حسین طهوریان",
				IsPublished: true,
				CreatedAt:   now.AddDate(0, 0, -3),
				UpdatedAt:   now.AddDate(0, 0, -3),
			},
		},
		services: []domain.Service{
			{
				Title:       "مشاوره مدیریت",
				Slug:        "management-consulting",
				Description: "مشاوره تخصصی در زمینه مدیریت کسب‌وکار و بهبود عملکرد",
				Content:     "",
				IsActive:    true,
			},
			{
				Title:       "توسعه کسب‌وکار",
				Slug:        "business-development",
				Description: "راهبری توسعه کسب‌وکار و ورود به بازارهای جدید",
				Content:     "",
				IsActive:    true,
			},
		},
		gallery: []domain.GalleryItem{
			{Title: "هلدینگ اول", ImageURL: "https://placehold.co/600x400/1A365D/C9A96E?text=Holding+1", Category: "هلدینگ", SortOrder: 1},
			{Title: "شرکت دوم", ImageURL: "https://placehold.co/600x400/1A365D/C9A96E?text=Company+2", Category: "شرکت", SortOrder: 2},
			{Title: "پروژه سوم", ImageURL: "https://placehold.co/600x400/1A365D/C9A96E?text=Project+3", Category: "پروژه", SortOrder: 3},
		},
		stats: []domain.Stat{
			{Label: "سال تجربه", Value: 20, Suffix: "+"},
			{Label: "پروژه موفق", Value: 150, Suffix: "+"},
			{Label: "مشتری راضی", Value: 85, Suffix: "+"},
			{Label: "تیم متخصص", Value: 45, Suffix: "+"},
		},
		contactInfo: &domain.ContactInfo{
			Address:   "تهران، خیابان ولیعصر",
			Phone:     "۰۲۱-۱۲۳۴۵۶۷۸",
			Email:     "info@tahoorian.ir",
			WorkHours: "شنبه تا چهارشنبه: ۹ صبح تا ۶ عصر",
		},
	}
}

func (r *Repository) GetAbout() (*domain.About, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.about == nil {
		return &domain.About{
			Title:   "درباره ما",
			Content: "دکتر حسین طهوریان با بیش از ۲۰ سال تجربه در مدیریت و راهبری کسب‌وکارها، همواره در مسیر تحول و توسعه کشور گام برداشته است.",
		}, nil
	}
	return r.about, nil
}

func (r *Repository) SaveAbout(about *domain.About) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if about.ID == 0 {
		about.ID = 1
	}
	r.about = about
	return nil
}

func (r *Repository) ListServices() ([]domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.services, nil
}

func (r *Repository) GetServiceBySlug(slug string) (*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.services {
		if s.Slug == slug {
			return &s, nil
		}
	}
	return nil, nil
}

func (r *Repository) CreateContact(contact *domain.Contact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	contact.ID = len(r.contacts) + 1
	contact.CreatedAt = time.Now()
	r.contacts = append(r.contacts, *contact)
	return nil
}

func (r *Repository) ListGallery(category string) ([]domain.GalleryItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if category == "" {
		return r.gallery, nil
	}
	var filtered []domain.GalleryItem
	for _, item := range r.gallery {
		if item.Category == category {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func (r *Repository) ListArticles() ([]domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	res := make([]domain.Article, len(r.articles))
	copy(res, r.articles)
	return res, nil
}

func (r *Repository) GetArticleByID(id int) (*domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.articles {
		if a.ID == id {
			copyArt := a
			return &copyArt, nil
		}
	}
	return nil, nil
}

func (r *Repository) GetArticleBySlug(slug string) (*domain.Article, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, a := range r.articles {
		if a.Slug == slug {
			copyArt := a
			return &copyArt, nil
		}
	}
	return nil, nil
}

func (r *Repository) CreateArticle(article *domain.Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	article.ID = r.nextArticleID
	r.nextArticleID++
	r.articles = append([]domain.Article{*article}, r.articles...)
	return nil
}

func (r *Repository) UpdateArticle(article *domain.Article) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, a := range r.articles {
		if a.ID == article.ID {
			r.articles[i] = *article
			return nil
		}
	}
	return nil
}

func (r *Repository) DeleteArticle(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, a := range r.articles {
		if a.ID == id {
			r.articles = append(r.articles[:i], r.articles[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *Repository) ListStats() ([]domain.Stat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stats, nil
}

func (r *Repository) GetContactInfo() (*domain.ContactInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.contactInfo, nil
}
