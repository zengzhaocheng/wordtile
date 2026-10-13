package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wordtile/backend/internal/auth"
	"github.com/wordtile/backend/internal/config"
	"github.com/wordtile/backend/internal/handler"
)

// Handlers 汇总全部 handler。
type Handlers struct {
	Auth      *handler.AuthHandler
	Deck      *handler.DeckHandler
	Community *handler.CommunityHandler
	Version   *handler.VersionHandler
	Game      *handler.GameHandler
	Stats     *handler.StatsHandler
}

func corsMiddleware(origins string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, o := range strings.Split(origins, ",") {
		allowed[strings.TrimSpace(o)] = true
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origins == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// New 构建 gin 引擎并注册全部路由。
func New(h *Handlers, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware(cfg.CORSOrigins))

	api := r.Group("/api/v1")

	/* ---------- 认证（限流） ---------- */
	authGroup := api.Group("/auth")
	authGroup.Use(auth.RateLimit(20, time.Minute))
	{
		authGroup.POST("/register", h.Auth.Register)
		authGroup.POST("/login", h.Auth.Login)
		authGroup.POST("/refresh", h.Auth.Refresh)
	}

	/* ---------- 在线题库市场（公开，无需登录） ---------- */
	api.GET("/marketplace/decks", h.Community.Marketplace)

	/* ---------- 需登录 ---------- */
	authed := api.Group("")
	authed.Use(auth.RequireAuth(cfg.JWTSecret))
	authed.GET("/users/me", h.Auth.Me)
	authed.PATCH("/users/me", h.Auth.UpdateMe)

	decks := authed.Group("/decks")
	{
		decks.GET("", h.Deck.List)
		decks.POST("", h.Deck.Create)
		decks.GET("/:id", h.Deck.GetDetail)
		decks.PATCH("/:id", h.Deck.Update)
		decks.DELETE("/:id", h.Deck.Delete)

		decks.GET("/:id/entries", h.Deck.ListEntries)
		decks.POST("/:id/entries", h.Deck.CreateEntry)

		decks.POST("/:id/publish", h.Community.Publish)
		decks.POST("/:id/fork", h.Community.Fork)
		decks.POST("/:id/like", h.Community.ToggleLike)
		decks.GET("/:id/comments", h.Community.ListComments)
		decks.POST("/:id/comments", h.Community.AddComment)

		decks.GET("/:id/commits", h.Version.ListCommits)
		decks.POST("/:id/commits", h.Version.CreateCommit)
		decks.POST("/:id/merge", h.Version.Merge)
		decks.GET("/:id/conflicts", h.Version.ListConflicts)
		decks.POST("/:id/conflicts/resolve", h.Version.ResolveConflict)
	}

	authed.PATCH("/entries/:id", h.Deck.UpdateEntry)
	authed.DELETE("/entries/:id", h.Deck.DeleteEntry)

	games := authed.Group("/games")
	{
		games.POST("/start", h.Game.Start)
		games.GET("/:id", h.Game.GetState)
		games.POST("/:id/match", h.Game.Match)
		games.POST("/:id/hint", h.Game.Hint)
		games.POST("/:id/finish", h.Game.Finish)
	}

	stats := authed.Group("/stats")
	{
		stats.GET("/contribution", h.Stats.Contribution)
		stats.GET("/decks/:id/progress", h.Stats.DeckProgress)
		stats.GET("/campaign", h.Stats.Campaign)
	}

	// 健康检查
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
