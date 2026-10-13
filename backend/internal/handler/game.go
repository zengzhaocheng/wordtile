package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/auth"
	"github.com/wordtile/backend/internal/service"
)

type GameHandler struct{ svc *service.GameService }

func NewGameHandler(svc *service.GameService) *GameHandler { return &GameHandler{svc: svc} }

func (h *GameHandler) Start(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	var req struct {
		DeckID     string   `json:"deck_id" binding:"required"`
		Mode       string   `json:"mode" binding:"required"`
		BlockCount int      `json:"block_count"`
		Level      int      `json:"level"`
		Tags       []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	deckID, err := uuid.Parse(req.DeckID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deck_id 格式不合法"})
		return
	}
	res, err := h.svc.Start(userID, deckID, req.Mode, req.Level, req.BlockCount, req.Tags)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *GameHandler) GetState(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	state, err := h.svc.GetState(userID, sessionID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *GameHandler) Match(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	var req struct {
		TermToken       string `json:"term_token" binding:"required"`
		DefinitionToken string `json:"definition_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	res, err := h.svc.Match(userID, sessionID, req.TermToken, req.DefinitionToken)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *GameHandler) Hint(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	res, err := h.svc.Hint(userID, sessionID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *GameHandler) Finish(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	var req struct {
		DurationMS int64 `json:"duration_ms"`
	}
	_ = c.ShouldBindJSON(&req)
	res, err := h.svc.Finish(userID, sessionID, req.DurationMS)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

type StatsHandler struct{ svc *service.StatsService }

func NewStatsHandler(svc *service.StatsService) *StatsHandler { return &StatsHandler{svc: svc} }

func (h *StatsHandler) Contribution(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	days, err := h.svc.Contribution(userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"days": days})
}

func (h *StatsHandler) DeckProgress(c *gin.Context) {
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
	points, err := h.svc.DeckProgress(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func (h *StatsHandler) Campaign(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	stats, err := h.svc.Campaign(userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"levels": stats})
}
