package utility

import (
	"time"

	"video-conference-be/internal/domain/user"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	Username string    `json:"username"`
	Role     user.Role `json:"role"`
	jwt.StandardClaims
}

func GenerateJWT(username string, role user.Role, ttl time.Duration) (string, error) {
	claims := &Claims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "vc-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(Config.JWTSecret))
}

func ParseJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(Config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
