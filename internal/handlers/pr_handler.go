package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreatePR создает новый PR и назначает ревьюверов
// @Summary Создать PR и автоматически назначить до 2 ревьюверов из команды автора
// @Description Создает новый PR и назначает до 2 активных ревьюверов из команды автора
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param input body CreatePRRequest true "Данные для создания PR"
// @Success 201 {object} Response{data=domain.PullRequest}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 409 {object} Response
// @Failure 500 {object} Response
// @Router /pullRequest/create [post]
func (h *Handler) CreatePR(c *gin.Context) {
	var req CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	pr, err := h.prService.CreatePR(c.Request.Context(), req.PRID, req.PRName, req.AuthorID)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusCreated, map[string]interface{}{
		"pr": pr,
	})
}

// MergePR помечает PR как MERGED
// @Summary Пометить PR как MERGED (идемпотентная операция)
// @Description Обновляет статус PR на MERGED. Операция идемпотентна.
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param input body MergePRRequest true "Данные для слияния PR"
// @Success 200 {object} Response{data=domain.PullRequest}
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /pullRequest/merge [post]
func (h *Handler) MergePR(c *gin.Context) {
	var req MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	pr, err := h.prService.MergePR(c.Request.Context(), req.PRID)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"pr": pr,
	})
}

// ReassignReviewer переназначает ревьювера на PR
// @Summary Переназначить конкретного ревьювера на другого из его команды
// @Description Заменяет одного ревьювера на случайного активного участника из той же команды
// @Tags PullRequests
// @Accept json
// @Produce json
// @Param input body ReassignReviewerRequest true "Данные для переназначения ревьювера"
// @Success 200 {object} Response{data=ReassignReviewerResponse}
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 409 {object} Response
// @Failure 500 {object} Response
// @Router /pullRequest/reassign [post]
func (h *Handler) ReassignReviewer(c *gin.Context) {
	var req ReassignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	newReviewerID, err := h.prService.ReassignReviewer(c.Request.Context(), req.PRID, req.OldReviewerID)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	// Получаем обновленный PR для ответа
	pr, err := h.prService.GetPR(c.Request.Context(), req.PRID)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}

// CreatePRRequest представляет запрос на создание PR
type CreatePRRequest struct {
	PRID     string `json:"pull_request_id" binding:"required"`
	PRName   string `json:"pull_request_name" binding:"required"`
	AuthorID string `json:"author_id" binding:"required"`
}

// MergePRRequest представляет запрос на слияние PR
type MergePRRequest struct {
	PRID string `json:"pull_request_id" binding:"required"`
}

// ReassignReviewerRequest представляет запрос на переназначение ревьювера
type ReassignReviewerRequest struct {
	PRID          string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_reviewer_id" binding:"required"`
}

