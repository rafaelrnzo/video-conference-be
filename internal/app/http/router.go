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
		AllowCredentials: true,
	}))

	userRepo := repository.NewUserRepository()
	groupRepo := repository.NewGroupRepository()
	recordRepo := repository.NewRecordRepository()

	authSvc := service.NewAuthService(userRepo)
	groupSvc := service.NewGroupService(groupRepo, userRepo)

	lkClient := utility.NewLivekitClient()
	lkSvc := service.NewLivekitService(lkClient)

	roomSvc := service.NewRoomService()
	userSvc := service.NewUserService()
	recordSvc := service.NewRecordService(recordRepo)

	authHandler := NewAuthHandler(authSvc)
	groupHandler := NewGroupHandler(groupSvc)
	lkHandler := NewLivekitHandler(lkSvc, roomSvc, groupSvc)
	roomHandler := NewRoomHandler(roomSvc)
	userHandler := NewUserHandler(userSvc)
	recordingHandler := NewRecordingHandler(lkSvc, recordSvc)

	// === PUBLIC ROUTES ===
	r.GET("/healthz", lkHandler.Health)
	r.GET("/public", authHandler.Public)
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/livekit/webhook", lkHandler.Webhook)

	api := r.Group("/api")
	api.Use(JWTAuthMiddleware())
	{
		api.GET("/protected", authHandler.Protected)
		api.POST("/livekit/token", lkHandler.GenerateToken)
		api.POST("/livekit/leave", lkHandler.LeaveRoom)
		api.POST("/livekit/kick", lkHandler.KickParticipant)

		api.GET("/rooms", roomHandler.ListRooms)
	}

	// === ADMIN ROUTES (/admin/...) ===
	adminGroup := r.Group("/admin")
	adminGroup.Use(JWTAuthMiddleware(), AdminOnly())
	{
		// GROUP MANAGEMENT
		adminGroup.GET("/groups", groupHandler.ListGroups)
		adminGroup.GET("/groups/:id", groupHandler.GetGroup)
		adminGroup.POST("/groups", groupHandler.CreateGroup)
		adminGroup.PATCH("/groups/:id", groupHandler.UpdateGroup)
		adminGroup.DELETE("/groups/:id", groupHandler.DeleteGroup)
		adminGroup.POST("/groups/:id/members", groupHandler.AddMember)
		adminGroup.DELETE("/groups/:id/members/:userId", groupHandler.RemoveMember)

		// LIVEKIT ROOMS
		adminGroup.GET("/livekit/rooms", lkHandler.ListActiveRooms)
		adminGroup.DELETE("/livekit/rooms/:name", lkHandler.DeleteActiveRoom)
		adminGroup.GET("/livekit/participants", lkHandler.ListParticipants)
		adminGroup.DELETE("/livekit/participants", lkHandler.RemoveParticipant)

		// RECORDINGS
		adminGroup.POST("/livekit/recordings/start", recordingHandler.StartRecording)
		adminGroup.POST("/livekit/recordings/stop", recordingHandler.StopRecording)
		adminGroup.GET("/recordings", recordingHandler.ListRecords)
		adminGroup.PATCH("/recordings/:id", recordingHandler.UpdateRecordName)
		adminGroup.DELETE("/recordings/:id", recordingHandler.DeleteRecord)

		// USERS
		adminGroup.GET("/users", userHandler.ListUsers)
		adminGroup.POST("/users", userHandler.CreateUser)
		adminGroup.PATCH("/users/:id", userHandler.UpdateUserRole)
		adminGroup.DELETE("/users/:id", userHandler.DeleteUser)

		// STATIC ROOMS (ADMIN MANAGE)
		adminGroup.GET("/rooms", roomHandler.ListRooms)
		adminGroup.POST("/rooms", roomHandler.CreateRoom)
		adminGroup.PATCH("/rooms/:id", roomHandler.UpdateRoom)
		adminGroup.DELETE("/rooms/:id", roomHandler.DeleteRoom)
	}

	return r
}
