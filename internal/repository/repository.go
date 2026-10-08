package repository

import (
	"context"
	"dr-tahoorian/internal/domain"
)

type Repository interface {
	GetAbout(ctx context.Context) (*domain.AboutInfo, error)
	GetContactInfo(ctx context.Context) (*domain.ContactInfo, error)
	GetStats(ctx context.Context) ([]domain.Stat, error)
	GetServices(ctx context.Context) ([]domain.Service, error)
	GetPackages(ctx context.Context) ([]domain.ServicePackage, error)
	GetGallery(ctx context.Context) ([]domain.GalleryItem, error)
	
	// Articles
	GetArticles(ctx context.Context, publishedOnly bool) ([]domain.Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (*domain.Article, error)
	GetArticleByID(ctx context.Context, id int64) (*domain.Article, error)
	CreateArticle(ctx context.Context, article *domain.Article) error
	UpdateArticle(ctx context.Context, article *domain.Article) error
	DeleteArticle(ctx context.Context, id int64) error

	// Categories
	GetCategories(ctx context.Context) ([]domain.ArticleCategory, error)
	CreateCategory(ctx context.Context, category *domain.ArticleCategory) error
	DeleteCategory(ctx context.Context, id int64) error

	// Contacts
	CreateContact(ctx context.Context, contact *domain.Contact) error
	GetContacts(ctx context.Context) ([]domain.Contact, error)
}
