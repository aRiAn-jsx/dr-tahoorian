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
