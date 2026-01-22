package rbac

import (
	"log"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

var Enforcer *casbin.Enforcer



// CheckPermission checks if a subject has permission on object for action
func CheckPermission(sub, obj, act string) (bool, error) {
	return Enforcer.Enforce(sub, obj, act)
}

// SystemPermission defines a permission structure
type SystemPermission struct {
	Label  string `json:"label"`
	Object string `json:"object"`
	Action string `json:"action"`
}

// SystemPermissions lists all available permissions in the system.
// This is used for seeding and for the frontend to know what can be assigned.
var SystemPermissions = []SystemPermission{
	{Label: "View Dashboard", Object: "dashboard", Action: "read"},
	
	{Label: "View Rooms", Object: "rooms", Action: "read"},
	{Label: "Create Rooms", Object: "rooms", Action: "create"},
	{Label: "Delete Rooms", Object: "rooms", Action: "delete"},
	{Label: "Manage Rooms", Object: "rooms", Action: "manage"}, // Covers kick/mute/etc if logic implies it

	{Label: "View Users", Object: "users", Action: "read"},
	{Label: "Manage Users", Object: "users", Action: "manage"},

	{Label: "View Recordings", Object: "recordings", Action: "read"},
	{Label: "Manage Recordings", Object: "recordings", Action: "manage"},

	{Label: "View Groups", Object: "groups", Action: "read"},
	{Label: "Manage Groups", Object: "groups", Action: "manage"},

	{Label: "Manage Roles", Object: "roles", Action: "manage"},
}

func InitEnforcer(db *gorm.DB) {
	// Initialize GORM adapter
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		log.Fatalf("failed to initialize casbin adapter: %v", err)
	}

	// Load model configuration
	e, err := casbin.NewEnforcer("internal/pkg/rbac/model.conf", adapter)
	if err != nil {
		log.Fatalf("failed to create casbin enforcer: %v", err)
	}

	// Load policies from DB
	err = e.LoadPolicy()
	if err != nil {
		log.Fatalf("failed to load policies: %v", err)
	}

	Enforcer = e
	
	// Seed default roles if they don't exist
	// Ensure 'admin' has ALL SystemPermissions
	for _, perm := range SystemPermissions {
		if !e.HasPolicy("admin", perm.Object, perm.Action) {
			e.AddPolicy("admin", perm.Object, perm.Action)
		}
	}

	// Check if 'user' has any policy
	if !e.HasPolicy("user", "dashboard", "read") {
		e.AddPolicy("user", "dashboard", "read")
	}

	log.Println("RBAC Enforcer initialized successfully")
}

// AddRoleForUser assigns a role to a user
func AddRoleForUser(user, role string) (bool, error) {
	return Enforcer.AddGroupingPolicy(user, role)
}

// AddPermissionForRole checking if it adds a policy
func AddPermissionForRole(role, obj, act string) (bool, error) {
	return Enforcer.AddPolicy(role, obj, act)
}
