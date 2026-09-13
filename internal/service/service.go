package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/tahoorian/tahoorian/internal/domain"
)

type Repository interface {
	GetAbout() (*domain.About, error)
	SaveAbout(about *domain.About) error
	ListServices() ([]domain.Service, error)
	GetServiceBySlug(slug string) (*domain.Service, error)
	CreateContact(contact *domain.Contact) error
	ListGallery(category string) ([]domain.GalleryItem, error)
	ListStats() ([]domain.Stat, error)
	GetContactInfo() (*domain.ContactInfo, error)
	ListArticles() ([]domain.Article, error)
	GetArticleByID(id int) (*domain.Article, error)
	GetArticleBySlug(slug string) (*domain.Article, error)
	CreateArticle(article *domain.Article) error
	UpdateArticle(article *domain.Article) error
	DeleteArticle(id int) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAbout() (*domain.About, error) {
	return s.repo.GetAbout()
}

func (s *Service) UpdateAbout(about *domain.About) error {
	about.UpdatedAt = time.Now()
	return s.repo.SaveAbout(about)
}

func (s *Service) ListServices() ([]domain.Service, error) {
	return s.repo.ListServices()
}

func (s *Service) GetService(slug string) (*domain.Service, error) {
	return s.repo.GetServiceBySlug(slug)
}

func (s *Service) SubmitContact(contact *domain.Contact) error {
	if contact.Name == "" || contact.Message == "" {
		return fmt.Errorf("name and message are required")
	}
	contact.CreatedAt = time.Now()
	slog.Info("new contact submission", "name", contact.Name, "email", contact.Email)
	return s.repo.CreateContact(contact)
}

func (s *Service) ListGallery(category string) ([]domain.GalleryItem, error) {
	return s.repo.ListGallery(category)
}

func (s *Service) ListStats() ([]domain.Stat, error) {
	return s.repo.ListStats()
}

func (s *Service) GetContactInfo() (*domain.ContactInfo, error) {
	return s.repo.GetContactInfo()
}


func (s *Service) ListArticles() ([]domain.Article, error) {
	return s.repo.ListArticles()
}

func (s *Service) ListPublishedArticles() ([]domain.Article, error) {
	articles, err := s.repo.ListArticles()
	if err != nil {
		return nil, err
	}
	var published []domain.Article
	for _, a := range articles {
		if a.IsPublished {
			published = append(published, a)
		}
	}
	return published, nil
}

func (s *Service) GetArticleByID(id int) (*domain.Article, error) {
	return s.repo.GetArticleByID(id)
}

func (s *Service) GetArticleBySlug(slug string) (*domain.Article, error) {
	return s.repo.GetArticleBySlug(slug)
}

func (s *Service) CreateArticle(article *domain.Article) error {
	if article.Title == "" {
		return fmt.Errorf("عنوان مقاله الزامی است")
	}
	if article.Slug == "" {
		article.Slug = fmt.Sprintf("article-%d", time.Now().Unix())
	}
	article.CreatedAt = time.Now()
	article.UpdatedAt = time.Now()
	return s.repo.CreateArticle(article)
}

func (s *Service) UpdateArticle(article *domain.Article) error {
	if article.ID <= 0 {
		return fmt.Errorf("شناسه مقاله نامعتبر است")
	}
	if article.Title == "" {
		return fmt.Errorf("عنوان مقاله الزامی است")
	}
	article.UpdatedAt = time.Now()
	return s.repo.UpdateArticle(article)
}

func (s *Service) DeleteArticle(id int) error {
	return s.repo.DeleteArticle(id)
}
