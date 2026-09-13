package service_test

import (
	"testing"

	"github.com/tahoorian/tahoorian/internal/domain"
	memrepo "github.com/tahoorian/tahoorian/internal/repository/memory"
	"github.com/tahoorian/tahoorian/internal/service"
)

func testSetup() (*service.Service, *memrepo.Repository) {
	repo := memrepo.New()
	svc := service.New(repo)
	return svc, repo
}

func TestArticles(t *testing.T) {
	svc, _ := testSetup()

	articles, err := svc.ListArticles()
	if err != nil {
		t.Fatalf("expected no error listing articles, got %v", err)
	}
	if len(articles) == 0 {
		t.Fatalf("expected initial articles, got 0")
	}

	// Create article
	newArt := &domain.Article{
		Title:       "تست مقاله جدید",
		Slug:        "test-new-article",
		Summary:     "خلاصه تست",
		Content:     "محتوای کامل مقاله تست",
		Author:      "نویسنده تست",
		IsPublished: true,
	}
	if err := svc.CreateArticle(newArt); err != nil {
		t.Fatalf("failed to create article: %v", err)
	}
	if newArt.ID == 0 {
		t.Errorf("expected article ID to be assigned")
	}

	// Get article by ID
	fetched, err := svc.GetArticleByID(newArt.ID)
	if err != nil || fetched == nil {
		t.Fatalf("failed to get article by ID: %v", err)
	}
	if fetched.Title != newArt.Title {
		t.Errorf("expected title %s, got %s", newArt.Title, fetched.Title)
	}

	// Get article by Slug
	fetchedBySlug, err := svc.GetArticleBySlug("test-new-article")
	if err != nil || fetchedBySlug == nil {
		t.Fatalf("failed to get article by slug: %v", err)
	}
	if fetchedBySlug.ID != newArt.ID {
		t.Errorf("expected ID %d, got %d", newArt.ID, fetchedBySlug.ID)
	}

	// Update article
	fetchedBySlug.Title = "عنوان به روز شده"
	if err := svc.UpdateArticle(fetchedBySlug); err != nil {
		t.Fatalf("failed to update article: %v", err)
	}

	updated, _ := svc.GetArticleByID(newArt.ID)
	if updated.Title != "عنوان به روز شده" {
		t.Errorf("expected updated title, got %s", updated.Title)
	}

	// List Published
	published, err := svc.ListPublishedArticles()
	if err != nil {
		t.Fatalf("failed to list published articles: %v", err)
	}
	if len(published) < 1 {
		t.Errorf("expected at least 1 published article")
	}

	// Delete article
	if err := svc.DeleteArticle(newArt.ID); err != nil {
		t.Fatalf("failed to delete article: %v", err)
	}
	deleted, _ := svc.GetArticleByID(newArt.ID)
	if deleted != nil {
		t.Errorf("expected article to be deleted")
	}
}

func TestContactValidation(t *testing.T) {
	svc, _ := testSetup()

	// Invalid contact submission (missing required fields)
	err := svc.SubmitContact(&domain.Contact{})
	if err == nil {
		t.Fatalf("expected error for empty contact, got nil")
	}

	// Valid contact submission
	validContact := &domain.Contact{
		Name:    "علی رضایی",
		Email:   "ali@example.com",
		Phone:   "09123456789",
		Subject: "ارتباط با ما",
		Message: "سلام، درخواست مشاوره دارم.",
	}
	if err := svc.SubmitContact(validContact); err != nil {
		t.Fatalf("expected successful contact submission, got %v", err)
	}
}

func TestServicesAndAbout(t *testing.T) {
	svc, _ := testSetup()

	about, err := svc.GetAbout()
	if err != nil || about == nil {
		t.Fatalf("failed to get about: %v", err)
	}

	about.Title = "درباره ما - به روز شده"
	if err := svc.UpdateAbout(about); err != nil {
		t.Fatalf("failed to update about: %v", err)
	}

	services, err := svc.ListServices()
	if err != nil || len(services) == 0 {
		t.Fatalf("failed to list services: %v", err)
	}
}
