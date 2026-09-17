package sqlite

import (
	"database/sql"
	"log/slog"
)

// countRows returns the number of rows in a table. It is used by Seed to make
// seeding idempotent so it never overwrites or duplicates existing content.
func countRows(db *sql.DB, table string) int {
	row := db.QueryRow(`SELECT COUNT(*) FROM ` + table)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0
	}
	return n
}

// Seed populates the database with default content (mirroring the in-memory
// repository) whenever the corresponding tables are empty. Migrations create
// the schema; Seed ensures the site renders meaningful content on first boot.
// It is safe to call on every start.
func Seed(db *sql.DB) error {
	// --- About ---
	if countRows(db, "about") == 0 {
		_, err := db.Exec(`INSERT INTO about (title, content, image_url) VALUES (?, ?, ?)`,
			"درباره ما",
			"دکتر حسین طهوریان با بیش از ۲۰ سال تجربه در مدیریت و راهبری کسب‌وکارها، همواره در مسیر تحول و توسعه کشور گام برداشته است.",
			"/static/images/magnific_background-remover_cpvKmif0eP.png")
		if err != nil {
			return err
		}
	}

	// --- Services ---
	if countRows(db, "services") == 0 {
		if _, err := db.Exec(`INSERT INTO services (title, slug, description, content, is_active) VALUES (?, ?, ?, ?, ?)`,
			"مشاوره مدیریت", "management-consulting",
			"مشاوره تخصصی در زمینه مدیریت کسب‌وکار و بهبود عملکرد", "", 1); err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO services (title, slug, description, content, is_active) VALUES (?, ?, ?, ?, ?)`,
			"توسعه کسب‌وکار", "business-development",
			"راهبری توسعه کسب‌وکار و ورود به بازارهای جدید", "", 1); err != nil {
			return err
		}
	}

	// --- Gallery ---
	if countRows(db, "gallery") == 0 {
		if _, err := db.Exec(`INSERT INTO gallery (title, image_url, category, sort_order) VALUES (?, ?, ?, ?)`,
			"هلدینگ اول", "https://placehold.co/600x400/1A365D/C9A96E?text=Holding+1", "هلدینگ", 1); err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO gallery (title, image_url, category, sort_order) VALUES (?, ?, ?, ?)`,
			"شرکت دوم", "https://placehold.co/600x400/1A365D/C9A96E?text=Company+2", "شرکت", 2); err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO gallery (title, image_url, category, sort_order) VALUES (?, ?, ?, ?)`,
			"پروژه سوم", "https://placehold.co/600x400/1A365D/C9A96E?text=Project+3", "پروژه", 3); err != nil {
			return err
		}
	}

	// --- Stats ---
	if countRows(db, "stats") == 0 {
		stats := map[string]int{
			"سال تجربه": 20,
			"پروژه موفق": 150,
			"مشتری راضی": 85,
			"تیم متخصص": 45,
		}
		for label, value := range stats {
			if _, err := db.Exec(`INSERT INTO stats (label, value, suffix) VALUES (?, ?, ?)`,
				label, value, "+"); err != nil {
				return err
			}
		}
	}

	// --- Contact info ---
	if countRows(db, "contact_info") == 0 {
		_, err := db.Exec(`INSERT INTO contact_info (address, phone, email, work_hours) VALUES (?, ?, ?, ?)`,
			"تهران، خیابان ولیعصر",
			"۰۲۱-۱۲۳۴۵۶۷۸",
			"info@tahoorian.ir",
			"شنبه تا چهارشنبه: ۹ صبح تا ۶ عصر")
		if err != nil {
			return err
		}
	}

	// --- Articles ---
	if countRows(db, "articles") == 0 {
		if _, err := db.Exec(`INSERT INTO articles (title, slug, summary, content, image_url, author, is_published) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"اصول بنیادین در مدیریت استراتژیک هلدینگ‌ها",
			"principles-of-strategic-holding-management",
			"چگونه شرکت‌های مادر و هلدینگ‌ها می‌توانند هم‌افزایی ارزش را بین شرکت‌های تابعه ایجاد و هدایت کنند.",
			"مدیریت استراتژیک در هلدینگ‌ها نیازمند تفکیک دقیق میان سطح استراتژی کسب‌وکار و سطح استراتژی شرکتی است.",
			"/static/images/کتاب-های-سایت-دکتر-طهوریان-1024x569.jpg",
			"دکتر حسین طهوریان", 1); err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO articles (title, slug, summary, content, image_url, author, is_published) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"تبدیل نوآوری به ثبت اختراع و محصول تجاری",
			"turning-innovation-into-patents-and-products",
			"مسیر تجاری‌سازی ایده‌ها و اختراعات از فرضیه تا تولید صنعتی و ورود به بازار رقابتی.",
			"ایده‌ها تا زمانی که به یک مدل قابل اتکا و تکرارپذیر تبدیل نشوند، صرفاً در حد پتانسیل باقی می‌مانند.",
			"/static/images/اختراعات-دکتر-سایت.png",
			"دکتر حسین طهوریان", 1); err != nil {
			return err
		}
		if _, err := db.Exec(`INSERT INTO articles (title, slug, summary, content, image_url, author, is_published) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"رهبری تیم‌های ارزش‌آفرین در شرایط ابهام",
			"leadership-in-uncertainty",
			"ابزارهای کلیدی یک مدیر در تصمیم‌گیری‌های حساس و هدایت انگیزه سازمان در شرایط نااطمینانی اقتصادی.",
			"رهبری در شرایط ابهام صرفاً پیش‌بینی دقیق آینده نیست؛ بلکه خلق قابلیت انطباق‌پذیری بالا در سازمان است.",
			"/static/images/شرکت.png",
			"دکتر حسین طهوریان", 1); err != nil {
			return err
		}
	}

	slog.Info("database seeded with default content")
	return nil
}