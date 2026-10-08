package domain

import "time"

type Article struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	Slug               string    `json:"slug"`
	Category           string    `json:"category"`
	Summary            string    `json:"summary"`
	Content            string    `json:"content"`
	ImageURL           string    `json:"image_url"`
	Author             string    `json:"author"`
	IsPublished        bool      `json:"is_published"`
	Keywords           string    `json:"keywords,omitempty"`
	ReadingTime        string    `json:"reading_time,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	FormattedDate      string    `json:"formatted_date,omitempty"`
	FormattedDateShort string    `json:"formatted_date_short,omitempty"`
}

type ArticleCategory struct {
	ID                 int64     `json:"id"`
	Slug               string    `json:"slug"`
	Title              string    `json:"title"`
	CreatedAt          time.Time `json:"created_at"`
	FormattedDateShort string    `json:"formatted_date_short,omitempty"`
}

type Service struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Content     string `json:"content"`
	ImageURL    string `json:"image_url"`
	IsActive    bool   `json:"is_active"`
}

type Stat struct {
	ID     int64  `json:"id"`
	Label  string `json:"label"`
	Value  int64  `json:"value"`
	Suffix string `json:"suffix"`
}

type GalleryItem struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	ImageURL  string `json:"image_url"`
	Category  string `json:"category"`
	SortOrder int    `json:"sort_order"`
}

type Contact struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type AboutInfo struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	ImageURL string `json:"image_url"`
}

type ContactInfo struct {
	Address   string `json:"address"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	WorkHours string `json:"work_hours"`
}

type ServicePackage struct {
	ID          string   `json:"id"`
	Theme       string   `json:"theme"`
	Title       string   `json:"title"`
	LatinTitle  string   `json:"latin_title"`
	Tag         string   `json:"tag"`
	Icon        string   `json:"icon"`
	IsFeatured  bool     `json:"is_featured"`
	Audience    string   `json:"audience"`
	Description string   `json:"description"`
	Features    []string `json:"features"`
	Deliverables []string `json:"deliverables"`
	ProcessSteps []ProcessStep `json:"process_steps"`
	Sessions    string   `json:"sessions"`
	Duration    string   `json:"duration"`
	Location    string   `json:"location"`
	Guarantee   string   `json:"guarantee"`
	ActionText  string   `json:"action_text"`
}

type ProcessStep struct {
	Num   string `json:"num"`
	Title string `json:"title"`
}
