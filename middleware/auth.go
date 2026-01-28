package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"video-conference-be/config"
	"video-conference-be/models"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	Provider *oidc.Provider
	Verifier *oidc.IDTokenVerifier
	Config   *config.Config
}

func NewAuthMiddleware(cfg *config.Config) (*AuthMiddleware, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, cfg.KeycloakURL+"/realms/"+cfg.Realm)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %v", err)
	}

	oidcConfig := &oidc.Config{
		ClientID: cfg.ClientID,
		// If you want to verify audience, uncomment the following line
		// ClientID: cfg.ClientID,
		SkipClientIDCheck: true, // Keycloak access tokens might not have aud matching client_id
	}
	verifier := provider.Verifier(oidcConfig)

	return &AuthMiddleware{
		Provider: provider,
		Verifier: verifier,
		Config:   cfg,
	}, nil
}

func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		ctx := context.Background()
		idToken, err := m.Verifier.Verify(ctx, parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			return
		}

		// Extract claims
		var claims struct {
			Sub               string                 `json:"sub"`
			Email             string                 `json:"email"`
			PreferredUsername string                 `json:"preferred_username"`
			RealmAccess       struct {
				Roles []string `json:"roles"`
			} `json:"realm_access"`
		}
		if err := idToken.Claims(&claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to parse claims"})
			return
		}

		// Determine Role (Simple mapping)
		role := "user"
		for _, r := range claims.RealmAccess.Roles {
			if r == "admin" {
				role = "admin"
				break
			}
		}

		// Set User in Context
		user := models.User{
			KeycloakID: claims.Sub,
			Email:      claims.Email,
			Name:       claims.PreferredUsername,
			Role:       role,
		}

		c.Set("user", user)
		c.Next()
	}
}

// GetUserFromContext helper to retrieve user from gin context
func GetUserFromContext(c *gin.Context) (models.User, error) {
	userInterface, exists := c.Get("user")
	if !exists {
		return models.User{}, errors.New("user not found in context")
	}
	user, ok := userInterface.(models.User)
	if !ok {
		return models.User{}, errors.New("invalid user type in context")
	}
	return user, nil
}
