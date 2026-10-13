package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wordtile/backend/internal/apperr"
)

func respondErr(c *gin.Context, err error) {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		c.JSON(ae.Status, ae)
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误: " + err.Error()})
}

func parseUUID(c *gin.Context, name string) (string, bool) {
	id := c.Param(name)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数 " + name})
		return "", false
	}
	return id, true
}
