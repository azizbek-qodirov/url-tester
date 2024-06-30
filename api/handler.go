package api

import (
	"net/http"

	"url-tester/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) PostTest(c *gin.Context) {
	var reqModels []*models.RequestModel
	if err := c.ShouldBindJSON(&reqModels); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res := performLoadTest(reqModels)

	c.JSON(http.StatusOK, res)
}
