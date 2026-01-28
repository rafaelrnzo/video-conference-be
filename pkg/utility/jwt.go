package utility

import (
	"log"
	"time"

	// "video-conference-be/internal/domain/user"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func GenerateJWT(username string, role string, ttl time.Duration) (string, error) {
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
	signed, err := token.SignedString([]byte(Config.JWTSecret))
	if err != nil {
		return "", err
	}

	log.Printf("[JWT] generated token for user=%s role=%s exp=%d\n",
		username, role, claims.ExpiresAt)

	return signed, nil
}

func ParseJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(Config.JWTSecret), nil
	})
	if err != nil {
		log.Println("[JWT] ParseWithClaims error:", err)
		return nil, err
	}
	if !token.Valid {
		log.Println("[JWT] token invalid (signature/claims)")
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
