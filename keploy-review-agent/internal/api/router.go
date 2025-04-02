package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/keploy/keploy-review-agent/internal/config"
	"github.com/keploy/keploy-review-agent/internal/event"
)

func NewRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	
	// Webhook handlers
	webhookHandler := event.NewWebhookHandler(cfg)
	
	// GitHub webhook
	r.POST("/webhook/github", webhookHandler.HandleGitHub)
	
	// GitLab webhook
	r.POST("/webhook/gitlab", webhookHandler.HandleGitLab)
	
	// API endpoints for manual triggers
	api := r.Group("/api")
	{
		// Trigger analysis manually
		api.POST("/analyze", webhookHandler.HandleManualAnalysis)
		
		// Get analysis results
		api.GET("/results/:id", func(c *gin.Context) {
			// TODO: Implement results retrieval
			c.JSON(http.StatusNotImplemented, gin.H{
				"status": "not implemented",
			})
		})
	}
	
	return r
}
