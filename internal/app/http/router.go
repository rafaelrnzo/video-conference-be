package http

import (
	"time"

	"video-conference-be/internal/app/repository"
	"video-conference-be/internal/app/service"
	"video-conference-be/pkg/utility"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// ==== CORS supaya Next.js (localhost:3000) bisa call backend (8080) ====
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // FE URL kamu
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// init infra
	userRepo := repository.NewUserRepository()
	authSvc := service.NewAuthService(userRepo)
	lkClient := utility.NewLivekitClient()
	lkSvc := service.NewLivekitService(lkClient)

	authHandler := NewAuthHandler(authSvc)
	lkHandler := NewLivekitHandler(lkSvc)

	// PUBLIC
	r.GET("/healthz", lkHandler.Health)
	r.GET("/public", authHandler.Public)
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// AUTHENTICATED
	api := r.Group("/api")
	api.Use(JWTAuthMiddleware())
	{
		api.GET("/protected", authHandler.Protected)
		api.POST("/livekit/token", lkHandler.GenerateToken)
	}

	// ADMIN
	adminGroup := r.Group("/admin")
	adminGroup.Use(JWTAuthMiddleware(), AdminOnly())
	{
		adminGroup.POST("/livekit/rooms", lkHandler.CreateRoom)
		adminGroup.GET("/livekit/rooms", lkHandler.ListRooms)
		adminGroup.DELETE("/livekit/rooms/:name", lkHandler.DeleteRoom)

		adminGroup.GET("/livekit/participants", lkHandler.ListParticipants)
		adminGroup.DELETE("/livekit/participants", lkHandler.RemoveParticipant)
	}

	return r
}
