package main

import (
	"fmt"
	"log"

	"video-conference-be/internal/app/http"
	"video-conference-be/internal/domain/user"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

func autoMigrate(db *gorm.DB) {
	db.AutoMigrate(&user.User{})
	// &room.Room{} kalau sudah punya
}

func main() {
	utility.LoadConfig()
	utility.InitDB()
	autoMigrate(utility.DB)

	r := http.NewRouter()

	addr := ":" + utility.Config.Port
	fmt.Printf("Server running at http://localhost%s\n", addr)
	log.Fatal(r.Run(addr))
}
