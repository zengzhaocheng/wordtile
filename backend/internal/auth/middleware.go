package auth

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/apperr"
)

// UserID 从 gin 上下文取出当前用户 ID。
func UserID(c *gin.Context) (uuid.UUID, error) {
	v, ok := c.Get("userID")
	if !ok {
		return uuid.Nil, apperr.ErrUnauthorized
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, apperr.ErrUnauthorized
	}
	return id, nil
}

// RequireAuth 校验 Bearer Access Token。
func RequireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apperr.ErrUnauthorized)
			return
		}
		claims, err := ParseToken(strings.TrimPrefix(header, "Bearer "), secret)
		if err != nil || claims.Type != TokenTypeAccess {
			c.AbortWithStatusJSON(http.StatusUnauthorized, apperr.ErrUnauthorized)
			return
		}
		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// RateLimit 进程内固定窗口限流（按客户端 IP）。
// 说明：单实例部署够用；多副本生产环境建议替换为 Redis 实现（见 README 取舍）。
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	type bucket struct {
		count int
		reset time.Time
	}
	var mu sync.Mutex
	buckets := map[string]*bucket{}
	return func(c *gin.Context) {
		key := c.ClientIP()
		mu.Lock()
		now := time.Now()
		b := buckets[key]
		if b == nil || now.After(b.reset) {
			b = &bucket{reset: now.Add(window)}
			buckets[key] = b
		}
		b.count++
		over := b.count > limit
		mu.Unlock()
		if over {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后再试"})
			return
		}
		c.Next()
	}
}
