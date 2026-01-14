package main

import (
	"fmt"
	"log"
	"video-conference-be/internal/domain/models"

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

	r := http.NewRouter()

	port := utility.Config.Port
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Server running on %s\n", addr)
	log.Fatal(r.Run(addr))
}
