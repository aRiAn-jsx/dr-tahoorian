package sqlite

import (
	"context"
	"database/sql"
	"dr-tahoorian/internal/domain"
	"fmt"
	"time"

	_ "github.com/glebarez/go-sqlite"
)

type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(dsn string) (*SQLiteRepo, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	repo := &SQLiteRepo{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}
	if err := repo.seed(); err != nil {
		return nil, fmt.Errorf("seed failed: %w", err)
	}

	return repo, nil
}

func (r *SQLiteRepo) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS about (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		image_url TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS contact_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		address TEXT NOT NULL,
		phone TEXT NOT NULL,
		email TEXT NOT NULL,
		work_hours TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS services (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		content TEXT NOT NULL,
		image_url TEXT NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT 1
	);
	CREATE TABLE IF NOT EXISTS stats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		label TEXT NOT NULL,
		value INTEGER NOT NULL,
		suffix TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS gallery (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		image_url TEXT NOT NULL,
		category TEXT NOT NULL,
		sort_order INTEGER NOT NULL DEFAULT 0
	);
	CREATE TABLE IF NOT EXISTS article_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		slug TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		slug TEXT NOT NULL UNIQUE,
		category TEXT NOT NULL,
		summary TEXT NOT NULL,
		content TEXT NOT NULL,
		image_url TEXT NOT NULL,
		author TEXT NOT NULL DEFAULT 'دکتر حسین طهوریان',
		is_published BOOLEAN NOT NULL DEFAULT 1,
		keywords TEXT,
		reading_time TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS contacts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		phone TEXT NOT NULL,
		subject TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := r.db.Exec(schema)
	return err
}

func (r *SQLiteRepo) seed() error {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM about").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	_, _ = r.db.Exec(`INSERT INTO about (title, content, image_url) VALUES (?, ?, ?)`,
		"درباره ما",
		"دکتر حسین طهوریان با بیش از ۲۰ سال تجربه در مدیریت و راهبری کسب‌وکارها، همواره در مسیر تحول و توسعه کشور گام برداشته است.",
		"/static/images/magnific_background-remover_cpvKmif0eP.png",
	)

	_, _ = r.db.Exec(`INSERT INTO contact_info (address, phone, email, work_hours) VALUES (?, ?, ?, ?)`,
		"خیابان ولیعصر، برج تجاری آرین، طبقه ۱۲",
		"۰۲۱-۲۲۰۱۴۵۶۷",
		"info@tahoorian.ir",
		"شنبه تا چهارشنبه ۹:۰۰ الی ۱۷:۰۰",
	)

	_, _ = r.db.Exec(`INSERT INTO services (title, slug, description, content, image_url, is_active) VALUES
		('مشاوره مدیریت', 'management-consulting', 'مشاوره تخصصی در زمینه مدیریت کسب‌وکار و بهبود عملکرد', '', '', 1),
		('توسعه کسب‌وکار', 'business-development', 'راهبری توسعه کسب‌وکار و ورود به بازارهای جدید', '', '', 1)`)

	_, _ = r.db.Exec(`INSERT INTO stats (label, value, suffix) VALUES
		('سال تجربه', 20, '+'),
		('پروژه موفق', 150, '+'),
		('مشتری راضی', 85, '+'),
		('تیم متخصص', 45, '+')`)

	_, _ = r.db.Exec(`INSERT INTO gallery (title, image_url, category, sort_order) VALUES
		('هلدینگ سرآمد سرمایه ایلیا', '/static/images/37-100x100.png', 'هلدینگ', 1),
		('گروه بین‌المللی توسعه پتنت و نوآوری', '/static/images/36-100x100.png', 'ثبت اختراع', 2),
		('صندوق جسورانه سرمایه‌گذاری سلامت', '/static/images/35-100x100.png', 'سرمایه‌گذاری', 3)`)

	_, _ = r.db.Exec(`INSERT INTO article_categories (slug, title) VALUES
		('health-medicine', 'سلامت و طب سنتی'),
		('patents-innovation', 'ثبت اختراع و نوآوری'),
		('holding-investment', 'سرمایه‌گذاری و هلدینگ'),
		('management', 'مدیریت و استراتژی')`)

	return nil
}

func (r *SQLiteRepo) GetAbout(ctx context.Context) (*domain.AboutInfo, error) {
	var a domain.AboutInfo
	err := r.db.QueryRowContext(ctx, "SELECT title, content, image_url FROM about LIMIT 1").Scan(&a.Title, &a.Content, &a.ImageURL)
	if err == sql.ErrNoRows {
		return &domain.AboutInfo{
			Title:    "درباره ما",
			Content:  "دکتر حسین طهوریان با بیش از ۲۰ سال تجربه در مدیریت و راهبری کسب‌وکارها، همواره در مسیر تحول و توسعه کشور گام برداشته است.",
			ImageURL: "/static/images/magnific_background-remover_cpvKmif0eP.png",
		}, nil
	}
	return &a, err
}

func (r *SQLiteRepo) GetContactInfo(ctx context.Context) (*domain.ContactInfo, error) {
	var c domain.ContactInfo
	err := r.db.QueryRowContext(ctx, "SELECT address, phone, email, work_hours FROM contact_info LIMIT 1").Scan(&c.Address, &c.Phone, &c.Email, &c.WorkHours)
	if err == sql.ErrNoRows {
		return &domain.ContactInfo{
			Address:   "خیابان ولیعصر، برج تجاری آرین، طبقه ۱۲",
			Phone:     "۰۲۱-۲۲۰۱۴۵۶۷",
			Email:     "info@tahoorian.ir",
			WorkHours: "شنبه تا چهارشنبه ۹:۰۰ الی ۱۷:۰۰",
		}, nil
	}
	return &c, err
}

func (r *SQLiteRepo) GetStats(ctx context.Context) ([]domain.Stat, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, label, value, suffix FROM stats ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Stat
	for rows.Next() {
		var s domain.Stat
		if err := rows.Scan(&s.ID, &s.Label, &s.Value, &s.Suffix); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *SQLiteRepo) GetServices(ctx context.Context) ([]domain.Service, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, slug, description, content, image_url, is_active FROM services WHERE is_active = 1 ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Service
	for rows.Next() {
		var s domain.Service
		if err := rows.Scan(&s.ID, &s.Title, &s.Slug, &s.Description, &s.Content, &s.ImageURL, &s.IsActive); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *SQLiteRepo) GetPackages(ctx context.Context) ([]domain.ServicePackage, error) {
	return []domain.ServicePackage{
		{
			ID:          "roghgiri-health",
			Theme:       "emerald",
			Title:       "سلامت، طب سنتی و رگ‌گیری تخصصی",
			LatinTitle:  "Executive Health, Roghgiri & Traditional Medicine",
			Tag:         "اصالت طب سنتی و درمان دستی",
			Icon:        "heart-pulse",
			IsFeatured:  false,
			Audience:    "مدیران ارشد، کارآفرینان و افرادی که با دردهای مزمن ستون فقرات، گرفتگی‌های گردن و کتف، استرس مداوم و خستگی شغلی مواجهند.",
			Description: "رویکردی کاملاً طبیعی، علمی و منطبق بر گنجینه طب سنتی ایرانی و فنون کهن رگ‌گیری.",
			Features: []string{
				"معاینه بالینی مزاج و ارزیابی عارضه‌یابانه ستون فقرات و نقاط ماشه‌ای درد",
				"جلسات رگ‌گیری تخصصی با دست توسط دکتر طهوریان بدون داروهای شیمیایی",
				"آزادسازی فوری اسپاسم‌های عمقی گردن، کتف، پشت، فیله‌های کمر و سیاتیک",
			},
			Deliverables: []string{
				"شناسنامه تشخیصی طبع و ساختار اسکلتی-عضلانی",
				"پروتکل اختصاصی اصلاح سبک زندگی و رژیم مزاجی",
			},
			Sessions:   "۴ الی ۶ جلسه حضوری در کلینیک تخصصی",
			Duration:   "هر جلسه ۵۰ الی ۶۰ دقیقه",
			Location:   "کلینیک تخصصی دکتر طهوریان",
			Guarantee:  "محرمانگی کامل، اصالت فنون درمانی و پایش هفتگی نتایج بالینی",
			ActionText: "رزرو جلسه ارزیابی و رگ‌گیری تخصصی",
		},
		{
			ID:          "holding-investment",
			Theme:       "blue",
			Title:       "سرمایه‌گذاری، هلدینگ و ادغام و تملک (M&A)",
			LatinTitle:  "Holding Strategy, VC Investment & M&A Advisory",
			Tag:         "خلق ارزش پایدار و ساختار سرمایه هوشمند",
			Icon:        "trending-up",
			IsFeatured:  true,
			Audience:    "صاحبان سرمایه، مالکان هلدینگ‌ها، سهام‌داران کلیدی و شرکت‌های پیشرو با درآمد سالانه بالا.",
			Description: "راهبری جامع تخصیص بهینه سرمایه، طراحی ساختارهای حاکمیت شرکتی هلدینگ‌ها و مدیریت معاملات ادغام و تملک.",
			Sessions:   "همراهی استراتژیک ۶ الی ۱۲ ماهه",
			Duration:   "جلسات هفتگی و ماهانه هیئت‌مدیره",
			Location:   "دفتر مرکزی هلدینگ / جلسات حضوری و بین‌المللی",
			Guarantee:  "تضمین محرمانگی در سطح استانداردهای بین‌المللی (Strict NDA)",
			ActionText: "درخواست جلسه محرمانه سرمایه‌گذاری",
		},
	}, nil
}

func (r *SQLiteRepo) GetGallery(ctx context.Context) ([]domain.GalleryItem, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, title, image_url, category, sort_order FROM gallery ORDER BY sort_order ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.GalleryItem
	for rows.Next() {
		var g domain.GalleryItem
		if err := rows.Scan(&g.ID, &g.Title, &g.ImageURL, &g.Category, &g.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, nil
}

func (r *SQLiteRepo) GetArticles(ctx context.Context, publishedOnly bool) ([]domain.Article, error) {
	query := "SELECT id, title, slug, category, summary, content, image_url, author, is_published, COALESCE(keywords, ''), COALESCE(reading_time, ''), created_at, updated_at FROM articles"
	if publishedOnly {
		query += " WHERE is_published = 1"
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Article
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(&a.ID, &a.Title, &a.Slug, &a.Category, &a.Summary, &a.Content, &a.ImageURL, &a.Author, &a.IsPublished, &a.Keywords, &a.ReadingTime, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.FormattedDate = a.CreatedAt.Format("2006/01/02")
		a.FormattedDateShort = a.CreatedAt.Format("2006/01/02")
		list = append(list, a)
	}
	return list, nil
}

func (r *SQLiteRepo) GetArticleBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	query := "SELECT id, title, slug, category, summary, content, image_url, author, is_published, COALESCE(keywords, ''), COALESCE(reading_time, ''), created_at, updated_at FROM articles WHERE slug = ?"
	var a domain.Article
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&a.ID, &a.Title, &a.Slug, &a.Category, &a.Summary, &a.Content, &a.ImageURL, &a.Author, &a.IsPublished, &a.Keywords, &a.ReadingTime, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	a.FormattedDate = a.CreatedAt.Format("2006/01/02")
	a.FormattedDateShort = a.CreatedAt.Format("2006/01/02")
	return &a, nil
}

func (r *SQLiteRepo) GetArticleByID(ctx context.Context, id int64) (*domain.Article, error) {
	query := "SELECT id, title, slug, category, summary, content, image_url, author, is_published, COALESCE(keywords, ''), COALESCE(reading_time, ''), created_at, updated_at FROM articles WHERE id = ?"
	var a domain.Article
	err := r.db.QueryRowContext(ctx, query, id).Scan(&a.ID, &a.Title, &a.Slug, &a.Category, &a.Summary, &a.Content, &a.ImageURL, &a.Author, &a.IsPublished, &a.Keywords, &a.ReadingTime, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	a.FormattedDate = a.CreatedAt.Format("2006/01/02")
	a.FormattedDateShort = a.CreatedAt.Format("2006/01/02")
	return &a, nil
}

func (r *SQLiteRepo) CreateArticle(ctx context.Context, a *domain.Article) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO articles (title, slug, category, summary, content, image_url, author, is_published, keywords, reading_time, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Title, a.Slug, a.Category, a.Summary, a.Content, a.ImageURL, a.Author, a.IsPublished, a.Keywords, a.ReadingTime, now, now,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		a.ID = id
	}
	return nil
}

func (r *SQLiteRepo) UpdateArticle(ctx context.Context, a *domain.Article) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE articles SET title = ?, slug = ?, category = ?, summary = ?, content = ?, image_url = ?, author = ?, is_published = ?, keywords = ?, reading_time = ?, updated_at = ? WHERE id = ?`,
		a.Title, a.Slug, a.Category, a.Summary, a.Content, a.ImageURL, a.Author, a.IsPublished, a.Keywords, a.ReadingTime, now, a.ID,
	)
	return err
}

func (r *SQLiteRepo) DeleteArticle(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM articles WHERE id = ?", id)
	return err
}

func (r *SQLiteRepo) GetCategories(ctx context.Context) ([]domain.ArticleCategory, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, slug, title, created_at FROM article_categories ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.ArticleCategory
	for rows.Next() {
		var c domain.ArticleCategory
		if err := rows.Scan(&c.ID, &c.Slug, &c.Title, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.FormattedDateShort = c.CreatedAt.Format("2006/01/02")
		list = append(list, c)
	}
	return list, nil
}

func (r *SQLiteRepo) CreateCategory(ctx context.Context, c *domain.ArticleCategory) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, "INSERT INTO article_categories (slug, title, created_at) VALUES (?, ?, ?)", c.Slug, c.Title, now)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		c.ID = id
	}
	return nil
}

func (r *SQLiteRepo) DeleteCategory(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM article_categories WHERE id = ?", id)
	return err
}

func (r *SQLiteRepo) CreateContact(ctx context.Context, c *domain.Contact) error {
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO contacts (name, email, phone, subject, message, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		c.Name, c.Email, c.Phone, c.Subject, c.Message, now,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		c.ID = id
	}
	return nil
}

func (r *SQLiteRepo) GetContacts(ctx context.Context) ([]domain.Contact, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, email, phone, subject, message, created_at FROM contacts ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Contact
	for rows.Next() {
		var c domain.Contact
		if err := rows.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Subject, &c.Message, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}
