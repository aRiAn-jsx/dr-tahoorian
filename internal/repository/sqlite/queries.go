package sqlite

import (
	"database/sql"
	"time"

	"github.com/tahoorian/tahoorian/internal/domain"
)

func (r *Repository) GetAbout() (*domain.About, error) {
	row := r.db.QueryRow(`SELECT id, title, content, image_url, created_at, updated_at FROM about LIMIT 1`)
	about := &domain.About{}
	err := row.Scan(&about.ID, &about.Title, &about.Content, &about.ImageURL, &about.CreatedAt, &about.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return about, err
}

func (r *Repository) SaveAbout(about *domain.About) error {
	if about.ID == 0 {
		_, err := r.db.Exec(`INSERT INTO about (title, content, image_url) VALUES (?, ?, ?)`, about.Title, about.Content, about.ImageURL)
		return err
	}
	_, err := r.db.Exec(`UPDATE about SET title = ?, content = ?, image_url = ?, updated_at = ? WHERE id = ?`, about.Title, about.Content, about.ImageURL, time.Now(), about.ID)
	return err
}

func (r *Repository) ListServices() ([]domain.Service, error) {
	rows, err := r.db.Query(`SELECT id, title, slug, description, content, image_url, is_active, created_at, updated_at FROM services WHERE is_active = 1 ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []domain.Service
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.Content, &s.ImageURL, &s.IsActive, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		services = append(services, s)
	}
	return services, nil
}

func (r *Repository) GetServiceBySlug(slug string) (*domain.Service, error) {
	row := r.db.QueryRow(`SELECT id, title, slug, description, content, image_url, is_active, created_at, updated_at FROM services WHERE slug = ? AND is_active = 1`, slug)
	s := &domain.Service{}
	err := row.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.Content, &s.ImageURL, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func (r *Repository) CreateContact(contact *domain.Contact) error {
	_, err := r.db.Exec(`INSERT INTO contacts (name, email, phone, subject, message) VALUES (?, ?, ?, ?, ?)`, contact.Name, contact.Email, contact.Phone, contact.Subject, contact.Message)
	return err
}

func (r *Repository) ListGallery(category string) ([]domain.GalleryItem, error) {
	var rows *sql.Rows
	var err error
	if category != "" {
		rows, err = r.db.Query(`SELECT id, title, image_url, category, sort_order, created_at FROM gallery WHERE category = ? ORDER BY sort_order ASC`, category)
	} else {
		rows, err = r.db.Query(`SELECT id, title, image_url, category, sort_order, created_at FROM gallery ORDER BY sort_order ASC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.GalleryItem
	for rows.Next() {
		var item domain.GalleryItem
		if err := rows.Scan(&item.ID, &item.Title, &item.ImageURL, &item.Category, &item.SortOrder, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) ListStats() ([]domain.Stat, error) {
	rows, err := r.db.Query(`SELECT id, label, value, suffix FROM stats ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []domain.Stat
	for rows.Next() {
		var s domain.Stat
		if err := rows.Scan(&s.ID, &s.Label, &s.Value, &s.Suffix); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *Repository) GetContactInfo() (*domain.ContactInfo, error) {
	row := r.db.QueryRow(`SELECT address, phone, email, work_hours FROM contact_info LIMIT 1`)
	info := &domain.ContactInfo{}
	err := row.Scan(&info.Address, &info.Phone, &info.Email, &info.WorkHours)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return info, err
}
