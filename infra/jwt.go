package infra

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"lyonbot.github.com/my_app/misc"
)

var jwtSecret = []byte(misc.Getenv("JWT_SECRET", "50450000-dead-beef-1234-7ee4f3e70000"))

type JWTClaims struct {
	jwt.RegisteredClaims // 用户ID在 Subject 字段中

	Role string `json:"role"` // 可能没有，可能 "anon" 或 "service_role"
}

// 生成 JWT token
func GenerateJWT(userId string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour)

	claims := JWTClaims{
		Role: "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			Issuer:    "my_app",
			Subject:   userId,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)

	return token, err
}

// 解析 JWT token
func ParseJWT(token string) (*JWTClaims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*JWTClaims); ok && tokenClaims.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
