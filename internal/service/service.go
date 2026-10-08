package service

import (
	"context"
	"dr-tahoorian/internal/domain"
	"dr-tahoorian/internal/repository"
	"fmt"
	"net/http"
	"net/url"
)

type Service struct {
	repo             repository.Repository
	telegramBotToken string
	telegramChatID   string
}

func NewService(repo repository.Repository, telegramBotToken, telegramChatID string) *Service {
	return &Service{
		repo:             repo,
		telegramBotToken: telegramBotToken,
		telegramChatID:   telegramChatID,
	}
}

func (s *Service) GetHomeData(ctx context.Context) (*domain.AboutInfo, []domain.Stat, []domain.Service, []domain.Article, []domain.GalleryItem, error) {
	about, _ := s.repo.GetAbout(ctx)
	stats, _ := s.repo.GetStats(ctx)
	services, _ := s.repo.GetServices(ctx)
	articles, _ := s.repo.GetArticles(ctx, true)
	gallery, _ := s.repo.GetGallery(ctx)
	return about, stats, services, articles, gallery, nil
}

func (s *Service) GetAboutData(ctx context.Context) (*domain.AboutInfo, error) {
	return s.repo.GetAbout(ctx)
}

func (s *Service) GetServicesData(ctx context.Context) ([]domain.Service, []domain.ServicePackage, error) {
	services, err := s.repo.GetServices(ctx)
	if err != nil {
		return nil, nil, err
	}
	packages, err := s.repo.GetPackages(ctx)
	return services, packages, err
}

func (s *Service) GetProjectsData(ctx context.Context) ([]domain.GalleryItem, error) {
	return s.repo.GetGallery(ctx)
}

func (s *Service) GetArticlesData(ctx context.Context) ([]domain.Article, []domain.ArticleCategory, error) {
	articles, err := s.repo.GetArticles(ctx, true)
	if err != nil {
		return nil, nil, err
	}
	categories, err := s.repo.GetCategories(ctx)
	return articles, categories, err
}

func (s *Service) GetArticleDetail(ctx context.Context, slug string) (*domain.Article, []domain.Article, error) {
	article, err := s.repo.GetArticleBySlug(ctx, slug)
	if err != nil {
		return nil, nil, err
	}
	all, _ := s.repo.GetArticles(ctx, true)
	var related []domain.Article
	for _, a := range all {
		if a.ID != article.ID {
			related = append(related, a)
		}
		if len(related) >= 3 {
			break
		}
	}
	return article, related, nil
}

func (s *Service) SubmitContact(ctx context.Context, c *domain.Contact) error {
	if err := s.repo.CreateContact(ctx, c); err != nil {
		return err
	}

	// Telegram notification in background
	if s.telegramBotToken != "" && s.telegramChatID != "" {
		go s.sendTelegramNotification(c)
	}
	return nil
}

func (s *Service) sendTelegramNotification(c *domain.Contact) {
	text := fmt.Sprintf("📩 پیام جدید از وب‌سایت دکتر طهوریان:\n\n👤 نام: %s\n📞 تلفن: %s\n✉️ ایمیل: %s\n📌 موضوع: %s\n\n📝 متن پیام:\n%s",
		c.Name, c.Phone, c.Email, c.Subject, c.Message)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.telegramBotToken)
	_, _ = http.PostForm(apiURL, url.Values{
		"chat_id": {s.telegramChatID},
		"text":    {text},
	})
}
