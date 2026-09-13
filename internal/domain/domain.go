package domain

import "time"

type Article struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Slug        string    `json:"slug" db:"slug"`
	Summary     string    `json:"summary" db:"summary"`
	Content     string    `json:"content" db:"content"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	Author      string    `json:"author" db:"author"`
	IsPublished bool      `json:"is_published" db:"is_published"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type About struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Service struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Slug        string    `json:"slug" db:"slug"`
	Description string    `json:"description" db:"description"`
	Content     string    `json:"content" db:"content"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Contact struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	Subject   string    `json:"subject" db:"subject"`
	Message   string    `json:"message" db:"message"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type GalleryItem struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	ImageURL  string    `json:"image_url" db:"image_url"`
	Category  string    `json:"category" db:"category"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Stat struct {
	ID    int    `json:"id" db:"id"`
	Label string `json:"label" db:"label"`
	Value int    `json:"value" db:"value"`
	Suffix string `json:"suffix" db:"suffix"`
}

type ContactInfo struct {
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	WorkHours string `json:"work_hours"`
}

