package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims 自定义 JWT 载荷。
type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Type   string    `json:"typ"`
	jwt.RegisteredClaims
}

// GenerateToken 签发 HS256 JWT。
func GenerateToken(userID uuid.UUID, tokenType, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    "wordtile",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// ParseToken 校验并解析 JWT。
func ParseToken(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
