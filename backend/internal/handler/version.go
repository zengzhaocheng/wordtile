package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/auth"
	"github.com/wordtile/backend/internal/model"
	"github.com/wordtile/backend/internal/service"
)

// VersionHandler commit / merge / conflicts（Git-like 版本模块）。
type VersionHandler struct {
	svc     *service.MergeService
	deckSvc *service.DeckService
}

func NewVersionHandler(svc *service.MergeService, deckSvc *service.DeckService) *VersionHandler {
	return &VersionHandler{svc: svc, deckSvc: deckSvc}
}

func (h *VersionHandler) ListCommits(c *gin.Context) {
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
	commits, err := h.deckSvc.ListCommits(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"commits": commits})
}

func (h *VersionHandler) CreateCommit(c *gin.Context) {
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
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	commit, err := h.svc.Commit(userID, deckID, req.Message)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"commit": commit})
}

func (h *VersionHandler) Merge(c *gin.Context) {
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
		SourceDeckID   string `json:"source_deck_id"`
		SourceCommitID string `json:"source_commit_id"`
	}
	_ = c.ShouldBindJSON(&req)

	sourceDeckID := uuid.Nil
	if req.SourceDeckID != "" {
		if sourceDeckID, err = uuid.Parse(req.SourceDeckID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "source_deck_id 格式不合法"})
			return
		}
	}
	res, err := h.svc.Merge(userID, deckID, sourceDeckID, req.SourceCommitID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *VersionHandler) ListConflicts(c *gin.Context) {
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
	status := c.DefaultQuery("status", "open")
	conflicts, err := h.svc.ListConflicts(userID, deckID, status)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"conflicts": conflicts})
}

func (h *VersionHandler) ResolveConflict(c *gin.Context) {
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
		ConflictID string `json:"conflict_id" binding:"required"`
		Resolution string `json:"resolution" binding:"required"` // ours / theirs / custom
		Custom     *model.EntrySnapshot `json:"custom"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	conflictID, err := uuid.Parse(req.ConflictID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conflict_id 格式不合法"})
		return
	}
	custom := model.EntrySnapshot{}
	if req.Custom != nil {
		custom = *req.Custom
	}
	res, err := h.svc.ResolveConflict(userID, deckID, conflictID, req.Resolution, custom)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}
