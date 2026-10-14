package handler

import (
	"dr-tahoorian/internal/domain"
	"html/template"
	"strings"
)

// funcMap holds the helpers the HTML templates in views/ rely on.
var funcMap = template.FuncMap{
	"hasPrefix": strings.HasPrefix,
	"add":       func(a, b int) int { return a + b },
	"join":      func(sep string, items []string) string { return strings.Join(items, sep) },

	// firstN returns at most n items from the list.
	"firstN": func(n int, items []string) []string {
		if len(items) > n {
			return items[:n]
		}
		return items
	},

	// before returns the text before the first occurrence of sep (the whole string if sep is absent).
	"before": func(s, sep string) string {
		if i := strings.Index(s, sep); i >= 0 {
			return s[:i]
		}
		return s
	},

	// truncate shortens s to n characters, adding "..." when it was cut.
	"truncate": func(s string, n int) string {
		r := []rune(s)
		if len(r) <= n {
			return s
		}
		return string(r[:n]) + "..."
	},

	// countCat counts the articles that belong to the category slug.
	"countCat": func(articles []domain.Article, slug string) int {
		n := 0
		for _, a := range articles {
			if a.Category == slug {
				n++
			}
		}
		return n
	},

	// catTitle resolves a category slug to its title (falls back to the slug).
	"catTitle": func(cats []domain.ArticleCategory, slug string) string {
		for _, c := range cats {
			if c.Slug == slug {
				return c.Title
			}
		}
		return slug
	},

	// safeHTML marks admin-authored article HTML as trusted.
	"safeHTML": func(s string) template.HTML { return template.HTML(s) },

	// tel turns a display phone number (possibly with Persian digits) into a dialable one.
	"tel": func(s string) string {
		var b strings.Builder
		for _, r := range s {
			switch {
			case r >= '0' && r <= '9', r == '+':
				b.WriteRune(r)
			case r >= '۰' && r <= '۹':
				b.WriteRune('0' + (r - '۰'))
			case r >= '٠' && r <= '٩':
				b.WriteRune('0' + (r - '٠'))
			}
		}
		return b.String()
	},
}
