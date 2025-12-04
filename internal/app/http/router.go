package http

import (
	"video-conference-be/internal/app/repository"
	"video-conference-be/internal/app/service"
	"video-conference-be/pkg/utility"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://192.168.100.130:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	userRepo := repository.NewUserRepository()
	authSvc := service.NewAuthService(userRepo)
	lkClient := utility.NewLivekitClient()
	lkSvc := service.NewLivekitService(lkClient)

	roomSvc := service.NewRoomService()

	authHandler := NewAuthHandler(authSvc)
	lkHandler := NewLivekitHandler(lkSvc, roomSvc)
	roomHandler := NewRoomHandler(roomSvc)

	userSvc := service.NewUserService()
	userHandler := NewUserHandler(userSvc)

	recordRepo := repository.NewRecordRepository()
	recordSvc := service.NewRecordService(recordRepo)
	recordingHandler := NewRecordingHandler(lkSvc, recordSvc)

	r.GET("/healthz", lkHandler.Health)
	r.GET("/public", authHandler.Public)
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	api := r.Group("/api")
	api.Use(JWTAuthMiddleware())
	{
		api.GET("/protected", authHandler.Protected)
		api.POST("/livekit/token", lkHandler.GenerateToken)
	}

	adminGroup := r.Group("/admin")
	adminGroup.Use(JWTAuthMiddleware(), AdminOnly())
	{
		adminGroup.GET("/livekit/rooms", lkHandler.ListActiveRooms)
		adminGroup.DELETE("/livekit/rooms/:name", lkHandler.DeleteActiveRoom)

		adminGroup.GET("/livekit/participants", lkHandler.ListParticipants)
		adminGroup.DELETE("/livekit/participants", lkHandler.RemoveParticipant)

		adminGroup.POST("/livekit/recordings/start", recordingHandler.StartRecording)
		adminGroup.POST("/livekit/recordings/stop", recordingHandler.StopRecording)

		adminGroup.GET("/recordings", recordingHandler.ListRecords)
		adminGroup.PATCH("/recordings/:id", recordingHandler.UpdateRecordName)
		adminGroup.DELETE("/recordings/:id", recordingHandler.DeleteRecord)

		adminGroup.GET("/users", userHandler.ListUsers)
		adminGroup.POST("/users", userHandler.CreateUser)
		adminGroup.PATCH("/users/:id", userHandler.UpdateUserRole)
		adminGroup.DELETE("/users/:id", userHandler.DeleteUser)

		adminGroup.GET("/rooms", roomHandler.ListRooms)
		adminGroup.POST("/rooms", roomHandler.CreateRoom)
		adminGroup.PATCH("/rooms/:id", roomHandler.UpdateRoom)
		adminGroup.DELETE("/rooms/:id", roomHandler.DeleteRoom)
	}

	return r
}
