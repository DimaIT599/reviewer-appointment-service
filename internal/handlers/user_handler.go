package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetIsActive устанавливает флаг активности пользователя
// @Summary Установить флаг активности пользователя
// @Description Устанавливает флаг is_active для указанного пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Param input body SetIsActiveRequest true "Данные для установки активности"
// @Success 200 {object} Response{data=domain.User}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /users/setIsActive [post]
func (h *Handler) SetIsActive(c *gin.Context) {
	var req SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	user, err := h.userService.SetIsActive(c.Request.Context(), req.UserID, req.IsActive)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

// GetUserReviewPRs возвращает список PR, где пользователь назначен ревьювером
// @Summary Получить PR'ы, где пользователь назначен ревьювером
// @Description Возвращает список PR, где указанный пользователь является ревьювером
// @Tags Users
// @Produce json
// @Param user_id query string true "ID пользователя"
// @Success 200 {object} Response{data=[]domain.PullRequest}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /users/getReview [get]
func (h *Handler) GetUserReviewPRs(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "MISSING_PARAM",
				Message: "user_id is required",
			},
		})
		return
	}

	prs, err := h.userService.GetUserReviewPRs(c.Request.Context(), userID)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}

// SetIsActiveRequest представляет запрос на установку флага активности пользователя
type SetIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive bool   `json:"is_active" binding:"required"`
}
