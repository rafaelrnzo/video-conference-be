package http

import (
	"video-conference-be/internal/app/repository"
	"video-conference-be/internal/app/service"
	"video-conference-be/internal/pkg/rbac"
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

	// Move roleSvc init up
	roleRepo := repository.NewRoleRepository()
	roleSvc := service.NewRoleService(roleRepo)

	authSvc := service.NewAuthService(userRepo, roleSvc)
	groupSvc := service.NewGroupService(groupRepo, userRepo)

	lkClient := utility.NewLivekitClient()
	lkSvc := service.NewLivekitService(lkClient)

	roomSvc := service.NewRoomService()
	// roleRepo/roleSvc moved up
	
	userSvc := service.NewUserService(roleSvc)
	recordSvc := service.NewRecordService(recordRepo)

	authHandler := NewAuthHandler(authSvc)
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
	r.POST("/livekit/webhook", lkHandler.Webhook)

	api := r.Group("/api")
	api.Use(JWTAuthMiddleware())
	{
		api.GET("/protected", authHandler.Protected)
		api.POST("/livekit/token", lkHandler.GenerateToken)
		api.POST("/livekit/leave", lkHandler.LeaveRoom)
		api.POST("/livekit/kick", lkHandler.KickParticipant)
		api.POST("/livekit/admit", lkHandler.AdmitParticipant)
		api.GET("/meta", utilHandler.GetLinkMeta)

		api.GET("/rooms", roomHandler.ListRooms)
	}

	// === ADMIN ROUTES (/admin/...) ===
	adminGroup := r.Group("/admin")
	adminGroup.Use(JWTAuthMiddleware()) 
	// Removed AdminOnly(), using granular RBAC middleware below
	{
		// GROUP MANAGEMENT
		adminGroup.GET("/groups", rbac.Middleware("groups", "read"), groupHandler.ListGroups)
		adminGroup.GET("/groups/:id", rbac.Middleware("groups", "read"), groupHandler.GetGroup)
		adminGroup.POST("/groups", rbac.Middleware("groups", "manage"), groupHandler.CreateGroup)
		adminGroup.PATCH("/groups/:id", rbac.Middleware("groups", "manage"), groupHandler.UpdateGroup)
		adminGroup.DELETE("/groups/:id", rbac.Middleware("groups", "manage"), groupHandler.DeleteGroup)
		adminGroup.POST("/groups/:id/members", rbac.Middleware("groups", "manage"), groupHandler.AddMember)
		adminGroup.DELETE("/groups/:id/members/:userId", rbac.Middleware("groups", "manage"), groupHandler.RemoveMember)

		// LIVEKIT ROOMS (Active)
		adminGroup.GET("/livekit/rooms", rbac.Middleware("rooms", "read"), lkHandler.ListActiveRooms)
		adminGroup.DELETE("/livekit/rooms/:name", rbac.Middleware("rooms", "manage"), lkHandler.DeleteActiveRoom)
		adminGroup.GET("/livekit/participants", rbac.Middleware("rooms", "read"), lkHandler.ListParticipants)
		adminGroup.DELETE("/livekit/participants", rbac.Middleware("rooms", "manage"), lkHandler.RemoveParticipant)
		adminGroup.POST("/livekit/rooms/mute-all", rbac.Middleware("rooms", "manage"), lkHandler.MuteAll)
		adminGroup.POST("/livekit/rooms/permissions", rbac.Middleware("rooms", "manage"), lkHandler.UpdateRoomPermissions)

		// RECORDINGS
		adminGroup.POST("/livekit/recordings/start", rbac.Middleware("recordings", "manage"), recordingHandler.StartRecording)
		adminGroup.POST("/livekit/recordings/stop", rbac.Middleware("recordings", "manage"), recordingHandler.StopRecording)
		adminGroup.POST("/recordings/sync", rbac.Middleware("recordings", "manage"), recordingHandler.Sync)
		adminGroup.GET("/recordings", rbac.Middleware("recordings", "read"), recordingHandler.ListRecords)
		adminGroup.PATCH("/recordings/:id", rbac.Middleware("recordings", "manage"), recordingHandler.UpdateRecordName)
		adminGroup.DELETE("/recordings/:id", rbac.Middleware("recordings", "manage"), recordingHandler.DeleteRecord)

		// USERS
		adminGroup.GET("/users", rbac.Middleware("users", "read"), userHandler.ListUsers)
		adminGroup.POST("/users", rbac.Middleware("users", "manage"), userHandler.CreateUser)
		adminGroup.PATCH("/users/:id", rbac.Middleware("users", "manage"), userHandler.UpdateUserRole)
		adminGroup.DELETE("/users/:id", rbac.Middleware("users", "manage"), userHandler.DeleteUser)

		// STATIC ROOMS (ADMIN MANAGE)
		// Assuming static rooms management requires 'manage' or specific 'create'/'delete' permissions if defined
		adminGroup.GET("/rooms", rbac.Middleware("rooms", "read"), roomHandler.ListRooms)
		adminGroup.POST("/rooms", rbac.Middleware("rooms", "create"), roomHandler.CreateRoom)
		adminGroup.PATCH("/rooms/:id", rbac.Middleware("rooms", "manage"), roomHandler.UpdateRoom)
		adminGroup.DELETE("/rooms/:id", rbac.Middleware("rooms", "delete"), roomHandler.DeleteRoom)

		// POLLS
		// Polls are part of a room activity, let's assume 'rooms:manage' for now or add a poll permission.
		// SystemPermissions didn't explicit polls, usually room admins manage polls.
		adminGroup.POST("/polls", rbac.Middleware("rooms", "manage"), pollHandler.SavePoll)
        
        // RBAC (ROLES)
		roleHandler := NewRoleHandler(roleSvc)
        adminGroup.GET("/roles", rbac.Middleware("roles", "manage"), roleHandler.ListRoles)
        adminGroup.POST("/roles", rbac.Middleware("roles", "manage"), roleHandler.CreateRole)
        adminGroup.GET("/roles/:role/permissions", rbac.Middleware("roles", "manage"), roleHandler.GetRolePermissions)
        adminGroup.POST("/roles/permissions", rbac.Middleware("roles", "manage"), roleHandler.AddPermission)
        adminGroup.DELETE("/roles/permissions", rbac.Middleware("roles", "manage"), roleHandler.RemovePermission)
        adminGroup.GET("/system/permissions", rbac.Middleware("roles", "manage"), roleHandler.ListSystemPermissions)
	}
    
	return r
}
