package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/memodb-io/Acontext/docs"
	"github.com/memodb-io/Acontext/internal/modules/handler"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// zapLoggerMiddleware
func zapLoggerMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		log.Sugar().Infow("HTTP",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", dur.String(),
			"clientIP", c.ClientIP(),
		)
	}
}

type RouterDeps struct {
	DB             *gorm.DB
	Log            *zap.Logger
	ProjectHandler *handler.ProjectHandler
	SpaceHandler   *handler.SpaceHandler
	SessionHandler *handler.SessionHandler
}

func NewRouter(d RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), zapLoggerMiddleware(d.Log))

	// health
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	// swagger
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")
	{
		space := v1.Group("/space")
		{
			space.GET("/status")

			space.POST("/", d.SpaceHandler.CreateSpace)
			space.DELETE("/:space_id", d.SpaceHandler.DeleteSpace)

			space.PUT("/:space_id/configs", d.SpaceHandler.UpdateConfigs)
			space.GET("/:space_id/configs")

			space.GET("/:space_id/semantic_answer")
			space.GET("/:space_id/semantic_global")
			space.GET("/:space_id/semantic_grep")

			page := space.Group("/:space_id/page")
			{
				page.POST("/")
				page.DELETE("/:page_id")

				page.GET("/:page_id/properties")
				page.PUT("/:page_id/properties")

				page.GET("/:page_id/children")
				page.PUT("/:page_id/children/move_pages")
				page.PUT("/:page_id/children/move_blocks")
			}

			block := space.Group("/:space_id/block")
			{
				block.POST("/")
				block.DELETE("/:block_id")

				block.GET("/:block_id/properties")
				block.PUT("/:block_id/properties")

				block.GET("/:block_id/children")
				block.PUT("/:block_id/children/move_blocks")
			}
		}

		session := v1.Group("/session")
		{
			session.POST("/")
			session.DELETE("/:session_id")

			session.PUT("/:session_id/configs")
			session.GET("/:session_id/configs")

			session.POST("/:session_id/connect_to_space")

			session.POST("/:session_id/messages")
			session.GET("/:session_id/messages/status")

			session.GET("/:session_id/session_scratchpad")
			session.GET("/:session_id/tasks")
		}
	}
	return r
}
