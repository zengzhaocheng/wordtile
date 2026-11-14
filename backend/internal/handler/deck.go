package handler

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/wordtile/backend/internal/auth"
	"github.com/wordtile/backend/internal/deckio"
	"github.com/wordtile/backend/internal/service"
)

type DeckHandler struct {
	svc   *service.DeckService
	cache *service.MarketCache
}

func NewDeckHandler(svc *service.DeckService, cache *service.MarketCache) *DeckHandler {
	return &DeckHandler{svc: svc, cache: cache}
}

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

func (h *DeckHandler) invalidateMarket() {
	if h.cache != nil {
		h.cache.Invalidate()
	}
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
	h.invalidateMarket() // visibility 变更会影响市场列表
	c.JSON(http.StatusOK, gin.H{"deck": deck})
}

// Delete 删除题库。remove_from_market=true（默认）彻底删除；false 时若题库已公开，
// 只从「我的题库」移除并保留市场公开版。已 fork 出去的题库始终是独立副本，不受影响。
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
	// 默认 true：旧客户端直接 DELETE 仍是「彻底删除」，保持向后兼容。
	removeFromMarket := true
	if raw := c.Query("remove_from_market"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "remove_from_market 必须是 true/false"})
			return
		}
		removeFromMarket = v
	}
	res, err := h.svc.DeleteDeck(userID, deckID, removeFromMarket)
	if err != nil {
		respondErr(c, err)
		return
	}
	h.invalidateMarket()
	c.JSON(http.StatusOK, res)
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

/* ---------- JSON 导入 / 导出 ---------- */

// readImportBody 读取并限制导入文件体积（8 MiB）。
func readImportBody(c *gin.Context) ([]byte, bool) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 8<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
		return nil, false
	}
	return body, true
}

// ExportDeck GET /decks/:id/export → 题库 JSON 文档。
func (h *DeckHandler) ExportDeck(c *gin.Context) {
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
	f, err := h.svc.ExportDeck(userID, deckID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\"wordtile-deck-"+deckID.String()[:8]+".json\"")
	c.JSON(http.StatusOK, f)
}

// ImportNewDeck POST /decks/import → 从 JSON 文档创建新题库。
func (h *DeckHandler) ImportNewDeck(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	body, ok := readImportBody(c)
	if !ok {
		return
	}
	f, err := deckio.Parse(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	deck, err := h.svc.ImportNewDeck(userID, f)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"deck": deck, "imported": len(f.Entries)})
}

// ImportIntoDeck POST /decks/:id/import → 导入词条到已有题库（append / replace）。
func (h *DeckHandler) ImportIntoDeck(c *gin.Context) {
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
	body, ok := readImportBody(c)
	if !ok {
		return
	}
	f, mode, err := deckio.ParseImport(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.svc.ImportIntoDeck(userID, deckID, f, mode == "replace")
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deck_id": deckID, "imported": n, "mode": mode})
}
