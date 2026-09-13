package sqlite_test

import (
	"database/sql"
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/tahoorian/tahoorian/internal/domain"
	"github.com/tahoorian/tahoorian/internal/repository/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *sqlite.Repository) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	if err := sqlite.Migrate(db); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	repo := sqlite.NewRepository(db)
	return db, repo
}

func TestSQLiteRepository(t *testing.T) {
	db, repo := setupTestDB(t)
	defer db.Close()

	// Test About
	about := &domain.About{
		Title:   "درباره تست",
		Content: "محتوای درباره تست",
	}
	if err := repo.SaveAbout(about); err != nil {
		t.Fatalf("failed to save about: %v", err)
	}
	fetchedAbout, err := repo.GetAbout()
	if err != nil || fetchedAbout == nil {
		t.Fatalf("failed to get about: %v", err)
	}
	if fetchedAbout.Title != about.Title {
		t.Errorf("expected title %s, got %s", about.Title, fetchedAbout.Title)
	}

	// Test Articles CRUD
	art := &domain.Article{
		Title:       "مقاله دیتابیس",
		Slug:        "db-article-test",
		Summary:     "خلاصه",
		Content:     "محتوا",
		Author:      "دکتر طهوریان",
		IsPublished: true,
	}
	if err := repo.CreateArticle(art); err != nil {
		t.Fatalf("failed to create article in sqlite: %v", err)
	}
	if art.ID == 0 {
		t.Fatalf("expected ID to be assigned by sqlite")
	}

	articles, err := repo.ListArticles()
	if err != nil || len(articles) != 1 {
		t.Fatalf("expected 1 article in sqlite, got %d", len(articles))
	}

	artBySlug, err := repo.GetArticleBySlug("db-article-test")
	if err != nil || artBySlug == nil {
		t.Fatalf("failed to get article by slug from sqlite")
	}

	artBySlug.Title = "عنوان به روز شد"
	if err := repo.UpdateArticle(artBySlug); err != nil {
		t.Fatalf("failed to update article in sqlite: %v", err)
	}

	artByID, _ := repo.GetArticleByID(art.ID)
	if artByID.Title != "عنوان به روز شد" {
		t.Errorf("expected updated title, got %s", artByID.Title)
	}

	if err := repo.DeleteArticle(art.ID); err != nil {
		t.Fatalf("failed to delete article from sqlite: %v", err)
	}
	artDeleted, _ := repo.GetArticleByID(art.ID)
	if artDeleted != nil {
		t.Errorf("expected article to be deleted from sqlite")
	}
}
