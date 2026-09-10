package memory

import (
	"sync"
	"time"

	"github.com/tahoorian/tahoorian/internal/domain"
)

type Repository struct {
	mu           sync.RWMutex
	about        *domain.About
	services     []domain.Service
	contacts     []domain.Contact
	gallery      []domain.GalleryItem
	stats        []domain.Stat
	contactInfo  *domain.ContactInfo
}

func New() *Repository {
	return &Repository{
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
