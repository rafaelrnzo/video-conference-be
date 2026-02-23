package main

import (
	"context"
	"fmt"
	"log"
	"video-conference-be/internal/domain/models"

    "video-conference-be/internal/app/repository"
    "video-conference-be/internal/app/service"
	"video-conference-be/internal/app/http"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) {
	for _, m := range models.Models {
		db.AutoMigrate(m)
	}
}

func main() {
	utility.LoadConfig()
	utility.InitDB()
	utility.InitRedis()
	autoMigrate(utility.DB)

    roleRepo := repository.NewRoleRepository()
    roleSvc := service.NewRoleService(roleRepo)
    if err := roleSvc.InitDefaultRoles(context.Background()); err != nil {
        log.Printf("Warning: Failed to init default roles: %v", err)
    }

	r := http.NewRouter()

	addr := ":" + utility.Config.Port
	fmt.Printf("Server running at http://localhost%s\n", addr)
	log.Fatal(r.Run(addr))
}
