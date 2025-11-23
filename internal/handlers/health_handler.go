package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck проверяет работоспособность сервиса
// @Summary Проверить работоспособность сервиса
// @Description Возвращает статус работы сервиса
// @Tags Health
// @Produce json
// @Success 200 {object} Response{data=HealthResponse}
// @Router /health [get]
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Data: HealthResponse{
			Status:  "ok",
			Version: "1.0.0",
		},
	})
}

// HealthResponse представляет ответ на запрос проверки здоровья
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}
