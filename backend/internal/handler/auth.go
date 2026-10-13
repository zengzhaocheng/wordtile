package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wordtile/backend/internal/auth"
	"github.com/wordtile/backend/internal/service"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=6"`
		DisplayName string `json:"display_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法：" + err.Error()})
		return
	}
	user, access, refresh, err := h.svc.Register(req.Username, req.Email, req.Password, req.DisplayName)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user, "access_token": access, "refresh_token": refresh})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Identifier string `json:"identifier" binding:"required"`
		Password   string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	user, access, refresh, err := h.svc.Login(req.Identifier, req.Password)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "access_token": access, "refresh_token": refresh})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	access, refresh, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": access, "refresh_token": refresh})
}

// Me 获取当前用户信息。
func (h *AuthHandler) Me(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	user, err := h.svc.GetUser(userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateMe 更新当前用户资料 / 主题偏好。
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID, err := auth.UserID(c)
	if err != nil {
		respondErr(c, err)
		return
	}
	var req struct {
		DisplayName     *string `json:"display_name"`
		AvatarURL       *string `json:"avatar_url"`
		ThemePreference *string `json:"theme_preference"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数不合法"})
		return
	}
	user, err := h.svc.UpdateProfile(userID, req.DisplayName, req.AvatarURL, req.ThemePreference)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
