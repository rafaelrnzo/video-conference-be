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
			"http://lantech-vicon-fe-oo4azl-0de674-167-86-106-191.traefik.me",
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
	roleRepo := repository.NewRoleRepository()

	roleSvc := service.NewRoleService(roleRepo)
	authSvc := service.NewAuthService(userRepo, roleSvc)
	groupSvc := service.NewGroupService(groupRepo, userRepo)

	lkClient := utility.NewLivekitClient()
	lkSvc := service.NewLivekitService(lkClient)

	roomSvc := service.NewRoomService()
	userSvc := service.NewUserService()
	recordSvc := service.NewRecordService(recordRepo)

	authHandler := NewAuthHandler(authSvc)
	roleHandler := NewRoleHandler(roleSvc)
	groupHandler := NewGroupHandler(groupSvc)
	lkHandler := NewLivekitHandler(lkSvc, roomSvc, groupSvc)
	roomHandler := NewRoomHandler(roomSvc)
	userHandler := NewUserHandler(userSvc)
	recordingSvc := service.NewRecordingService(recordSvc)
	recordingHandler := NewRecordingHandler(lkSvc, recordSvc, recordingSvc)

	pollRepo := repository.NewPollRepository()
	pollSvc := service.NewPollService(pollRepo)
	pollHandler := NewPollHandler(pollSvc)

	utilHandler := NewUtilityHandler()

	// === PUBLIC ROUTES ===
	r.GET("/healthz", lkHandler.Health)
	r.GET("/public", authHandler.Public)
	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)
	r.POST("/sso-login", authHandler.SSOLogin)
	r.POST("/livekit/webhook", lkHandler.Webhook)
	r.GET("/public/rooms/:code", lkHandler.GetPublicRoom)
	r.POST("/public/join", lkHandler.JoinPublicRoom)

	api := r.Group("/api")
	api.Use(JWTAuthMiddleware())
	{
		api.GET("/me", authHandler.Me)
		api.GET("/protected", authHandler.Protected)
		api.POST("/livekit/token", lkHandler.GenerateToken)
		api.POST("/livekit/leave", lkHandler.LeaveRoom)
		api.POST("/livekit/kick", lkHandler.KickParticipant)
		api.POST("/livekit/ban", lkHandler.BanParticipant)
		api.POST("/livekit/unban", lkHandler.UnbanParticipant)
		api.POST("/livekit/admit", lkHandler.AdmitParticipant)
		api.GET("/meta", utilHandler.GetLinkMeta)

		api.GET("/rooms", roomHandler.ListRooms)
		api.GET("/rooms/:code", roomHandler.GetRoomByCode)
		api.GET("/presentations/:id", roomHandler.ProxyPresentation)
	}

	// === ADMIN ROUTES - Migrating to specific permissions ===
	// Note: We keep /admin prefix but check specific permissions.
	// AdminOnly() is removed from the group and applied per requirement or we use RequirePermission

	adminGroup := r.Group("/admin")
	adminGroup.Use(JWTAuthMiddleware())
	// adminGroup.Use(AdminOnly()) // REMOVED rigid check, using permission checks below
	{
		// GROUP MANAGEMENT
		adminGroup.GET("/groups", RequirePermission("group:manage"), groupHandler.ListGroups)
		adminGroup.GET("/groups/:id", RequirePermission("group:manage"), groupHandler.GetGroup)
		adminGroup.POST("/groups", RequirePermission("group:manage"), groupHandler.CreateGroup)
		adminGroup.PATCH("/groups/:id", RequirePermission("group:manage"), groupHandler.UpdateGroup)
		adminGroup.DELETE("/groups/:id", RequirePermission("group:manage"), groupHandler.DeleteGroup)
		adminGroup.POST("/groups/:id/members", RequirePermission("group:manage"), groupHandler.AddMember)
		adminGroup.DELETE("/groups/:id/members/:userId", RequirePermission("group:manage"), groupHandler.RemoveMember)

		// LIVEKIT ROOMS (Real-time)
		// Assuming permission "room:read" / "room:update" / "room:delete"
		adminGroup.GET("/livekit/rooms", RequirePermission("room:read"), lkHandler.ListActiveRooms)
		adminGroup.DELETE("/livekit/rooms/:name", RequirePermission("room:delete"), lkHandler.DeleteActiveRoom) // Closing room

		// Participants
		adminGroup.GET("/livekit/participants", RequirePermission("room:read"), lkHandler.ListParticipants)       // or explicit participant permission
		adminGroup.DELETE("/livekit/participants", RequirePermission("room:update"), lkHandler.RemoveParticipant) // kicking is managing room
		adminGroup.POST("/livekit/rooms/mute-all", RequirePermission("room:update"), lkHandler.MuteAll)
		adminGroup.POST("/livekit/participants/mute", RequirePermission("room:update"), lkHandler.MuteParticipant)
		adminGroup.POST("/livekit/rooms/permissions", RequirePermission("room:update"), lkHandler.UpdateRoomPermissions)

		// RECORDINGS
		adminGroup.POST("/livekit/recordings/start", RequirePermission("recording:create"), recordingHandler.StartRecording) // if missing, maybe room:update?
		adminGroup.POST("/livekit/recordings/stop", RequirePermission("recording:create"), recordingHandler.StopRecording)
		adminGroup.POST("/recordings/sync", RequirePermission("recording:create"), recordingHandler.Sync)
		adminGroup.GET("/recordings", RequirePermission("recording:read"), recordingHandler.ListRecords)
		adminGroup.PATCH("/recordings/:id", RequirePermission("recording:update"), recordingHandler.UpdateRecordName)
		adminGroup.DELETE("/recordings/:id", RequirePermission("recording:delete"), recordingHandler.DeleteRecord)

		// USERS
		adminGroup.GET("/users", RequirePermission("user:read"), userHandler.ListUsers)
		adminGroup.POST("/users", RequirePermission("user:create"), userHandler.CreateUser)
		adminGroup.PATCH("/users/:id", RequirePermission("user:update"), userHandler.UpdateUserRole)
		adminGroup.DELETE("/users/:id", RequirePermission("user:delete"), userHandler.DeleteUser)

		// STATIC ROOMS (DB)
		adminGroup.GET("/rooms", RequirePermission("room:read"), roomHandler.ListRooms)
		adminGroup.POST("/rooms", RequirePermission("room:create"), roomHandler.CreateRoom)

		adminGroup.PATCH("/rooms/:id", RequirePermission("room:update"), roomHandler.UpdateRoom)
		adminGroup.DELETE("/rooms/:id", RequirePermission("room:delete"), roomHandler.DeleteRoom)
		adminGroup.POST("/rooms/:id/presentation", RequirePermission("room:update"), roomHandler.UploadPresentation)

		adminGroup.GET("/presentations/:id", roomHandler.ProxyPresentation)

		// POLLS
		adminGroup.POST("/polls", RequirePermission("room:update"), pollHandler.SavePoll) // Polls are part of room activity

		// DYNAMIC ROLES - ConfigureRoleRoutes is separate, let's inline or check it
		// ConfigureRoleRoutes(adminGroup, roleHandler) <-- checks what?

		// Explicit Role Management Routes
		adminGroup.GET("/roles", RequirePermission("role:read"), roleHandler.ListRoles)
		adminGroup.POST("/roles", RequirePermission("role:create"), roleHandler.CreateRole)
		adminGroup.PATCH("/roles/:id", RequirePermission("role:update"), roleHandler.UpdateRole)
		adminGroup.DELETE("/roles/:id", RequirePermission("role:delete"), roleHandler.DeleteRole)
		adminGroup.GET("/roles/permissions", RequirePermission("role:read"), roleHandler.ListPermissions)
		adminGroup.POST("/roles/:id/permissions", RequirePermission("role:update"), roleHandler.AssignPermission)
		adminGroup.DELETE("/roles/:id/permissions/:permID", RequirePermission("role:update"), roleHandler.RevokePermission)
	}

	return r
}
