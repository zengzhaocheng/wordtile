package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/auth"
	"github.com/wordtile/backend/internal/service"
)

type DeckHandler struct{ svc *service.DeckService }

func NewDeckHandler(svc *service.DeckService) *DeckHandler { return &DeckHandler{svc: svc} }

func (h *DeckHandler) List(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	decks, err := h.svc.ListMyDecks(userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"decks": decks})
}

func (h *DeckHandler) Create(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	deck, err := h.svc.CreateDeck(userID, req.Name, req.Description)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"deck": deck})
}

func (h *DeckHandler) GetDetail(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	deckID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	detail, err := h.svc.GetDeckDetail(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *DeckHandler) Update(c *gin.Context) {
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
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Visibility  *string `json:"visibility"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	deck, err := h.svc.UpdateDeck(userID, deckID, req.Name, req.Description, req.Visibility)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deck": deck})
}

func (h *DeckHandler) Delete(c *gin.Context) {
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
	if err := h.svc.DeleteDeck(userID, deckID); err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

/* ---------- 词条 ---------- */

func (h *DeckHandler) ListEntries(c *gin.Context) {
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
	detail, err := h.svc.GetDeckDetail(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": detail.Entries})
}

func (h *DeckHandler) CreateEntry(c *gin.Context) {
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
		Term       string   `json:"term" binding:"required"`
		Definition string   `json:"definition" binding:"required"`
		Tags       []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	entry, err := h.svc.CreateEntry(userID, deckID, req.Term, req.Definition, req.Tags)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"entry": entry})
}

func (h *DeckHandler) UpdateEntry(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	entryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	var req struct {
		Term       *string  `json:"term"`
		Definition *string  `json:"definition"`
		Tags       []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	entry, err := h.svc.UpdateEntry(userID, entryID, req.Term, req.Definition, req.Tags, req.Tags != nil)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entry": entry})
}

func (h *DeckHandler) DeleteEntry(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	entryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 格式不合法"})
		return
	}
	if err := h.svc.DeleteEntry(userID, entryID); err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
