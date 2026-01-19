package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"shortlink/internal/models"
	"shortlink/internal/service"
)

// LinkHandler handles HTTP requests for link operations
type LinkHandler struct {
	service service.LinkService
	baseURL string
}

// NewLinkHandler creates a new link handler instance
func NewLinkHandler(svc service.LinkService, baseURL string) *LinkHandler {
	return &LinkHandler{
		service: svc,
		baseURL: baseURL,
	}
}

// CreateShortLink handles POST /api/shorten
func (h *LinkHandler) CreateShortLink(c *gin.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid request: URL is required",
		})
		return
	}

	link, err := h.service.CreateShortLink(c.Request.Context(), req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidURL) {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "Invalid URL format. URL must start with http:// or https://",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to create short link",
		})
		return
	}

	c.JSON(http.StatusCreated, models.ShortenResponse{
		ShortCode:   link.ShortCode,
		ShortURL:    h.baseURL + "/" + link.ShortCode,
		OriginalURL: link.OriginalURL,
	})
}

// Redirect handles GET /:code - redirects to original URL
func (h *LinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	// Skip common paths
	if code == "favicon.ico" || code == "api" || code == "health" {
		c.Status(http.StatusNotFound)
		return
	}

	originalURL, err := h.service.GetOriginalURL(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Short link not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve URL",
		})
		return
	}

	c.Redirect(http.StatusMovedPermanently, originalURL)
}

// GetLinkInfo handles GET /api/info/:code
func (h *LinkHandler) GetLinkInfo(c *gin.Context) {
	code := c.Param("code")

	link, err := h.service.GetLinkInfo(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Short link not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve link info",
		})
		return
	}

	c.JSON(http.StatusOK, models.LinkInfo{
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
		CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Clicks:      link.Clicks,
	})
}

// GetAllLinks handles GET /api/links
func (h *LinkHandler) GetAllLinks(c *gin.Context) {
	links, err := h.service.GetAllLinks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to retrieve links",
		})
		return
	}

	// Convert to LinkInfo slice
	var result []models.LinkInfo
	for _, link := range links {
		result = append(result, models.LinkInfo{
			ShortCode:   link.ShortCode,
			OriginalURL: link.OriginalURL,
			CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Clicks:      link.Clicks,
		})
	}

	if result == nil {
		result = []models.LinkInfo{}
	}

	c.JSON(http.StatusOK, result)
}

// DeleteLink handles DELETE /api/links/:code
func (h *LinkHandler) DeleteLink(c *gin.Context) {
	code := c.Param("code")

	err := h.service.DeleteLink(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, service.ErrLinkNotFound) {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: "Short link not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Failed to delete link",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Link deleted successfully",
	})
}
