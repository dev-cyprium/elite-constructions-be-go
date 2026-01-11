package http

import (
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/config"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/http/handlers"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.ErrorHandlerMiddleware())

	// Serve static files
	router.Static("/storage/img", "./storage/public/img")

	// Public routes (no auth)
	router.GET("/ping", handlers.Ping)

	public := router.Group("/api/pub")
	{
		public.GET("/projects", handlers.GetPublicProjects)
		public.GET("/projects/highlighted", handlers.GetHighlightedProjects)
		public.GET("/async/projects/page", handlers.GetPublicProjectsPaginated)
		public.GET("/projects/:id", handlers.GetPublicProject)
		public.GET("/testimonials", handlers.GetPublicTestimonials)
		public.POST("/testimonials", handlers.CreatePublicTestimonial)
		public.GET("/static-texts", handlers.GetPublicStaticTexts)
		public.GET("/configs", handlers.GetPublicConfigs)
		public.POST("/visitor-messages", handlers.CreateVisitorMessage)
	}

	// Auth routes
	auth := router.Group("/api")
	{
		auth.POST("/login", handlers.Login(cfg))
		auth.POST("/password-reset/complete", handlers.CompletePasswordReset(cfg))

		// Protected routes
		auth.Use(middleware.AuthMiddleware(cfg))
		{
			auth.POST("/logout", handlers.Logout)
			auth.GET("/me", handlers.GetMe)
		}
	}

	// Admin routes (require auth)
	admin := router.Group("/api")
	admin.Use(middleware.AuthMiddleware(cfg))
	{
		// Projects
		admin.GET("/projects", handlers.GetProjects)
		admin.GET("/projects/:id", handlers.GetProject)
		admin.POST("/projects", handlers.CreateProject(cfg))
		admin.PUT("/projects/:id", handlers.UpdateProject(cfg))
		admin.PUT("/projects/:id/highlight/toggle", handlers.ToggleHighlight)
		admin.DELETE("/projects/:id", handlers.DeleteProject(cfg))

		// Project Images
		admin.GET("/project-images/:id", handlers.GetProjectImage)

		// Testimonials
		admin.GET("/testimonials", handlers.GetTestimonials)
		admin.GET("/testimonials/:id", handlers.GetTestimonial)
		admin.POST("/testimonials", handlers.CreateTestimonial)
		admin.PUT("/testimonials/:id", handlers.UpdateTestimonial)
		admin.DELETE("/testimonials/:id", handlers.DeleteTestimonial)

		// Users
		admin.GET("/users", handlers.GetUsers)
		admin.GET("/users/:id", handlers.GetUser)
		admin.POST("/users", handlers.CreateUser)
		admin.PUT("/users/:id", handlers.UpdateUser)
		admin.DELETE("/users/:id", handlers.DeleteUser)

		// Static Texts
		admin.GET("/static-texts", handlers.GetStaticTexts)
		admin.GET("/static-texts/:id", handlers.GetStaticText)
		admin.PUT("/static-texts/:id", handlers.UpdateStaticText)

		// Configurations
		admin.PUT("/configs/:key", handlers.UpdateConfig)

		// Visitor Messages
		admin.GET("/visitor-messages", handlers.GetVisitorMessages)
		admin.DELETE("/visitor-messages/:id", handlers.DeleteVisitorMessage)
	}

	return router
}
