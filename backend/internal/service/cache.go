package service

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/wordtile/backend/internal/repository"
)

var cacheCtx = context.Background()

// MarketCache 市场列表缓存：Redis 优先（多实例共享），不可用时退化为进程内缓存。
type MarketCache struct {
	redis  *redis.Client
	ttl    time.Duration
	mu     sync.RWMutex
	memory map[string][]byte
}

func NewMarketCache(redisAddr string) *MarketCache {
	c := &MarketCache{ttl: 30 * time.Second, memory: map[string][]byte{}}
	if redisAddr != "" {
		c.redis = redis.NewClient(&redis.Options{Addr: redisAddr})
		// ping 失败不致命，Get 时自然回退内存
	}
	return c
}

func cacheKey(search, sort string) string {
	return "marketplace:" + search + "|" + sort
}

func (c *MarketCache) Get(search, sort string) ([]byte, bool) {
	key := cacheKey(search, sort)
	if c.redis != nil {
		if val, err := c.redis.Get(cacheCtx, key).Result(); err == nil {
			return []byte(val), true
		}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.memory[key]; ok {
		return v, true
	}
	return nil, false
}

func (c *MarketCache) Set(search, sort string, payload []byte) {
	key := cacheKey(search, sort)
	if c.redis != nil {
		if err := c.redis.Set(cacheCtx, key, string(payload), c.ttl).Err(); err == nil {
			return
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memory[key] = payload
	if len(c.memory) > 128 { // 简单防膨胀
		c.memory = map[string][]byte{}
	}
}

func (c *MarketCache) Invalidate() {
	c.mu.Lock()
	c.memory = map[string][]byte{}
	c.mu.Unlock()
	if c.redis != nil {
		keys := c.redis.Keys(cacheCtx, "marketplace:*").Val()
		if len(keys) > 0 {
			_ = c.redis.Del(cacheCtx, keys...).Err()
		}
	}
}

// MarshalDecks 序列化市场列表响应体。
func MarshalDecks(decks []repository.DeckWithAuthor) ([]byte, error) {
	return json.Marshal(map[string]any{"decks": decks})
}
