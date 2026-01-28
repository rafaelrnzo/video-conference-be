package models

type User struct {
	ID        string `json:"id" gorm:"primaryKey"` // UUID
	KeycloakID string `json:"keycloak_id" gorm:"uniqueIndex"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Role      string `json:"role"` // admin, user
}
