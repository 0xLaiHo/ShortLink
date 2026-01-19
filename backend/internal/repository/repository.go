package repository

import (
	"context"
	"time"

	"shortlink/internal/models"
)

// LinkRepository defines the interface for link storage operations
type LinkRepository interface {
	// Save stores a new link
	Save(ctx context.Context, link *models.Link) error

	// FindByCode retrieves a link by its short code
	FindByCode(ctx context.Context, code string) (*models.Link, error)

	// Exists checks if a short code already exists
	Exists(ctx context.Context, code string) (bool, error)

	// IncrementClicks increments the click counter for a link
	IncrementClicks(ctx context.Context, code string) error

	// FindAll retrieves all links
	FindAll(ctx context.Context) ([]*models.Link, error)

	// Delete removes a link by its short code
	Delete(ctx context.Context, code string) error
}

// LinkData represents the data structure stored in repository
type LinkData struct {
	OriginalURL string
	CreatedAt   time.Time
	Clicks      int64
}
