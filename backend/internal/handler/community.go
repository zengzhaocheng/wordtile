package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/auth"
	"github.com/wordtile/backend/internal/service"
)

// CommunityHandler 市场 / 发布 / fork / 点赞 / 评论。
type CommunityHandler struct {
	svc   *service.CommunityService
	cache *service.MarketCache
}

func NewCommunityHandler(svc *service.CommunityService, cache *service.MarketCache) *CommunityHandler {
	return &CommunityHandler{svc: svc, cache: cache}
}

func (h *CommunityHandler) Marketplace(c *gin.Context) {
	search := c.Query("search")
	sortKey := c.Query("sort")
	if key := c.Query("sort_by"); key != "" {
		sortKey = key
	}

	// 命中缓存直接返回（Redis 优先，进程内兜底）
	if payload, ok := h.cache.Get(search, sortKey); ok {
		c.Data(http.StatusOK, "application/json; charset=utf-8", payload)
		return
	}

	decks, err := h.svc.Marketplace(search, sortKey)
	if err != nil {
		respondErr(c, err)
		return
	}
	payload, err := service.MarshalDecks(decks)
	if err != nil {
		respondErr(c, err)
		return
	}
	h.cache.Set(search, sortKey, payload)
	c.Data(http.StatusOK, "application/json; charset=utf-8", payload)
}

func (h *CommunityHandler) Publish(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	deckID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	deck, err := h.svc.Publish(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	h.cache.Invalidate()
	c.JSON(http.StatusOK, gin.H{"deck": deck})
}

func (h *CommunityHandler) Fork(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	deckID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	fork, err := h.svc.Fork(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	h.cache.Invalidate()
	c.JSON(http.StatusCreated, gin.H{"deck": fork})
}

func (h *CommunityHandler) ToggleLike(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	deckID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	liked, count, err := h.svc.ToggleLike(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	h.cache.Invalidate()
	c.JSON(http.StatusOK, gin.H{"liked": liked, "likes_count": count})
}

func (h *CommunityHandler) ListComments(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	deckID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	comments, err := h.svc.ListComments(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments})
}

func (h *CommunityHandler) AddComment(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	deckID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	comment, err := h.svc.AddComment(userID, deckID, req.Content)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}
