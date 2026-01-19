package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shortlink/internal/handler"
	"shortlink/internal/middleware"
)

// Setup configures and returns the Gin router
func Setup(linkHandler *handler.LinkHandler, mode string) *gin.Engine {
	// Set Gin mode
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Apply global middleware
	r.Use(middleware.CORS())

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api")
	{
		api.POST("/shorten", linkHandler.CreateShortLink)
		api.GET("/info/:code", linkHandler.GetLinkInfo)
		api.GET("/links", linkHandler.GetAllLinks)
		api.DELETE("/links/:code", linkHandler.DeleteLink)
	}

	// Redirect endpoint - must be at root level
	r.GET("/:code", linkHandler.Redirect)

	return r
}
