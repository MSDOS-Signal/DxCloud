// Package jwt 提供 Access Token 的签发与校验（HS256）。
// Refresh Token 为不透明随机串，存 Redis（见 internal/iam），不走 JWT。
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uint64 `json:"uid"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Generate 签发 Access Token，返回 token 与 jti（用于黑名单）。
func Generate(secret string, ttl time.Duration, userID uint64, username string) (token string, jti string, err error) {
	jti = uuid.NewString()
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Issuer:    "dxcloud",
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString([]byte(secret))
	return token, jti, err
}

// Parse 校验签名与有效期，返回 Claims。
func Parse(secret, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
