package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/memodb-io/Acontext/docs"
	"github.com/memodb-io/Acontext/internal/handler"
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
			space.POST("/", d.SpaceHandler.CreateSpace)
			space.DELETE("/:space_id", d.SpaceHandler.DeleteSpace)
		}
	}
	return r
}
