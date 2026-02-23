package main

import (
	"context"
	"log"
	"time"

    "video-conference-be/internal/app/repository"
    "video-conference-be/internal/app/service"
	"video-conference-be/internal/domain/group"
	"video-conference-be/internal/domain/room"
	"video-conference-be/internal/domain/user"
	"video-conference-be/pkg/utility"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

func main() {
	// Load config and connect to DB
	utility.LoadConfig()
	utility.InitDB()

	db := utility.DB

	log.Println("Starting database seeder...")

    // Initialize Roles & Permissions
    roleRepo := repository.NewRoleRepository()
    roleSvc := service.NewRoleService(roleRepo)
    
    // We need a context
    ctx := context.Background()
    
    if err := roleSvc.InitDefaultRoles(ctx); err != nil {
        log.Printf("Failed to init default roles: %v\n", err)
    } else {
        log.Println("Default roles and permissions initialized.")
    }

	seedUsers(db)
	seedGroups(db)
	seedRooms(db)

	log.Println("Database seeding completed successfully.")
}

func seedUsers(db *gorm.DB) {
	log.Println("Seeding users...")

	password, _ := utility.HashPassword("password123")
	    
    var rAdmin, rUser struct { ID uint } 
    
    db.Table("roles").Where("name = ?", "admin").Select("id").Scan(&rAdmin)
    db.Table("roles").Where("name = ?", "user").Select("id").Scan(&rUser)
    
	users := []user.User{
		{
			Username:     "admin",
			PasswordHash: password,
			RoleID:       rAdmin.ID,
		},
		{
			Username:     "user1",
			PasswordHash: password,
			RoleID:       rUser.ID,
		},
		{
			Username:     "user2",
			PasswordHash: password,
			RoleID:       rUser.ID,
		},
	}

	for _, u := range users {
		var existing user.User
		if err := db.Where("username = ?", u.Username).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&u).Error; err != nil {
					log.Printf("Failed to create user %s: %v\n", u.Username, err)
				} else {
					log.Printf("User %s created.\n", u.Username)
				}
			} else {
				log.Printf("Error checking user %s: %v\n", u.Username, err)
			}
		} else {
             if existing.RoleID == 0 {
                  existing.RoleID = u.RoleID
                  db.Save(&existing)
                  log.Printf("Updated role for user %s\n", u.Username)
             } else {
			      log.Printf("User %s already exists.\n", u.Username)
             }
		}
	}
}

func seedGroups(db *gorm.DB) {
	log.Println("Seeding groups...")

	var user1, user2 user.User
	db.Where("username = ?", "user1").First(&user1)
	db.Where("username = ?", "user2").First(&user2)

	groups := []group.Group{
		{
			Name:        "Engineering",
			Description: "Engineering team group",
		},
		{
			Name:        "Product",
			Description: "Product team group",
		},
	}

	for _, g := range groups {
		var existing group.Group
		if err := db.Where("name = ?", g.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&g).Error; err != nil {
					log.Printf("Failed to create group %s: %v\n", g.Name, err)
				} else {
					log.Printf("Group %s created.\n", g.Name)
					
					if user1.ID != 0 {
						if err := db.Model(&g).Association("Members").Append(&user1); err != nil {
                            log.Printf("Failed to add user1 to group %s: %v\n", g.Name, err)
                        }
					}
                    if user2.ID != 0 {
                        if err := db.Model(&g).Association("Members").Append(&user2); err != nil {
                            log.Printf("Failed to add user2 to group %s: %v\n", g.Name, err)
                        }
                    }
				}
			} else {
				log.Printf("Error checking group %s: %v\n", g.Name, err)
			}
		} else {
			log.Printf("Group %s already exists.\n", g.Name)
		}
	}
}

func seedRooms(db *gorm.DB) {
	log.Println("Seeding rooms...")

    var engGroup group.Group
    db.Where("name = ?", "Engineering").First(&engGroup)
    
    var admin user.User
    db.Where("username = ?", "admin").First(&admin)

    var groupID *uint
    if engGroup.ID != 0 {
        id := engGroup.ID
        groupID = &id
    }

    rooms := []room.Room{
        {
            Name:            "Daily Standup",
            RoomCode:        "daily-standup",
            Description:     "Daily engineering standup",
            MaxParticipants: 20,
            StartDate:       time.Now(),
            EndDate:         time.Now().Add(1 * time.Hour),
            GroupID:         groupID,
            CreatedByID:     admin.ID,
            AssignedTo:      pq.StringArray{"user1"},
        },
        {
            Name:            "Product Review",
            RoomCode:        "product-review",
            Description:     "Weekly product review",
            MaxParticipants: 10,
            StartDate:       time.Now().Add(24 * time.Hour),
            EndDate:         time.Now().Add(25 * time.Hour),
            CreatedByID:     admin.ID,
        },
    }

	for _, r := range rooms {
		var existing room.Room
		if err := db.Where("room_code = ?", r.RoomCode).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&r).Error; err != nil {
					log.Printf("Failed to create room %s: %v\n", r.Name, err)
				} else {
					log.Printf("Room %s created.\n", r.Name)
				}
			} else {
				log.Printf("Error checking room %s: %v\n", r.Name, err)
			}
		} else {
			log.Printf("Room %s already exists.\n", r.Name)
		}
	}
}
