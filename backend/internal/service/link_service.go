package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"shortlink/internal/models"
	"shortlink/internal/repository"
)

// Common service errors
var (
	ErrInvalidURL     = errors.New("invalid URL format")
	ErrLinkNotFound   = errors.New("short link not found")
	ErrGenerateCode   = errors.New("failed to generate short code")
	ErrStoreFailed    = errors.New("failed to store link")
	ErrRetrieveFailed = errors.New("failed to retrieve link")
)

// LinkService defines the interface for link business operations
type LinkService interface {
	// CreateShortLink creates a new short link
	CreateShortLink(ctx context.Context, originalURL string) (*models.Link, error)

	// GetOriginalURL retrieves the original URL and increments click count
	GetOriginalURL(ctx context.Context, code string) (string, error)

	// GetLinkInfo retrieves information about a link
	GetLinkInfo(ctx context.Context, code string) (*models.Link, error)

	// GetAllLinks retrieves all links
	GetAllLinks(ctx context.Context) ([]*models.Link, error)

	// DeleteLink deletes a link
	DeleteLink(ctx context.Context, code string) error
}

// linkService implements LinkService
type linkService struct {
	repo repository.LinkRepository
}

// NewLinkService creates a new link service instance
func NewLinkService(repo repository.LinkRepository) LinkService {
	return &linkService{repo: repo}
}

// CreateShortLink creates a new short link
func (s *linkService) CreateShortLink(ctx context.Context, originalURL string) (*models.Link, error) {
	// Validate URL
	if err := validateURL(originalURL); err != nil {
		return nil, err
	}

	// Generate unique short code
	var shortCode string
	var err error
	for i := 0; i < 10; i++ {
		shortCode, err = generateShortCode()
		if err != nil {
			return nil, ErrGenerateCode
		}

		// Check if code already exists
		exists, _ := s.repo.Exists(ctx, shortCode)
		if !exists {
			break
		}
	}

	// Create link entity
	link := &models.Link{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		Clicks:      0,
	}

	// Save to repository
	if err := s.repo.Save(ctx, link); err != nil {
		return nil, ErrStoreFailed
	}

	return link, nil
}

// GetOriginalURL retrieves the original URL and increments click count
func (s *linkService) GetOriginalURL(ctx context.Context, code string) (string, error) {
	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrLinkNotFound) {
			return "", ErrLinkNotFound
		}
		return "", ErrRetrieveFailed
	}

	// Increment click counter asynchronously
	go func() {
		_ = s.repo.IncrementClicks(context.Background(), code)
	}()

	return link.OriginalURL, nil
}

// GetLinkInfo retrieves information about a link
func (s *linkService) GetLinkInfo(ctx context.Context, code string) (*models.Link, error) {
	link, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, repository.ErrLinkNotFound) {
			return nil, ErrLinkNotFound
		}
		return nil, ErrRetrieveFailed
	}
	return link, nil
}

// GetAllLinks retrieves all links
func (s *linkService) GetAllLinks(ctx context.Context) ([]*models.Link, error) {
	return s.repo.FindAll(ctx)
}

// DeleteLink deletes a link
func (s *linkService) DeleteLink(ctx context.Context, code string) error {
	err := s.repo.Delete(ctx, code)
	if errors.Is(err, repository.ErrLinkNotFound) {
		return ErrLinkNotFound
	}
	return err
}

// validateURL validates the URL format
func validateURL(url string) error {
	if len(url) < 10 {
		return ErrInvalidURL
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return ErrInvalidURL
	}
	return nil
}

// generateShortCode generates a random short code
func generateShortCode() (string, error) {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Use URL-safe base64 encoding and take first 6 characters
	code := base64.URLEncoding.EncodeToString(bytes)[:6]
	return code, nil
}
