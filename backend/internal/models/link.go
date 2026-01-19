package models

import "time"

// Link represents a shortened URL entity
type Link struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	Clicks      int64     `json:"clicks"`
}

// ShortenRequest represents the request body for creating a short link
type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

// ShortenResponse represents the response for a created short link
type ShortenResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// LinkInfo represents information about a short link (for API response)
type LinkInfo struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
	Clicks      int64  `json:"clicks"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string `json:"message"`
}
