package main

import (
	"log"
	"time"
	"video-conference-be/config"
	"video-conference-be/handlers"
	"video-conference-be/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize OIDC Middleware
	authMiddleware, err := middleware.NewAuthMiddleware(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize auth middleware: %v", err)
	}

	// 3. Initialize Router
	r := gin.Default()

	// 4. Setup CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 5. Initialize Handlers
	userHandler := handlers.NewUserHandler()

	// 6. Setup Routes
	api := r.Group("/api")
	{
		// Public Routes (if any)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Protected Routes
		protected := api.Group("/")
		protected.Use(authMiddleware.Handle())
		{
			protected.GET("/profile", userHandler.GetProfile)
			
			// Admin Routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RequireRole("admin"))
			{
				admin.GET("/", userHandler.AdminEndpoint)
			}
		}
	}

	// 7. Start Server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
