package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStats возвращает статистику по сервису
// @Summary Получить статистику по сервису
// @Description Возвращает различную статистику по работе сервиса
// @Tags Stats
// @Produce json
// @Success 200 {object} Response{data=StatsResponse}
// @Router /stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	totalPRs, err := h.statsRepo.GetTotalPRs(ctx)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	totalUsers, err := h.statsRepo.GetTotalUsers(ctx)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	activeUsers, err := h.statsRepo.GetActiveUsers(ctx)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	prsByStatus, err := h.statsRepo.GetPRsByStatus(ctx)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	topReviewers, err := h.statsRepo.GetTopReviewers(ctx, 10)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	topReviewersData := make([]map[string]interface{}, len(topReviewers))
	for i, reviewer := range topReviewers {
		topReviewersData[i] = map[string]interface{}{
			"user_id":      reviewer.UserID,
			"username":     reviewer.Username,
			"review_count": reviewer.ReviewCount,
		}
	}

	c.JSON(http.StatusOK, Response{
		Data: StatsResponse{
			TotalPRs:     totalPRs,
			TotalUsers:   totalUsers,
			PRsByStatus:  prsByStatus,
			ActiveUsers:  activeUsers,
			TopReviewers: topReviewersData,
		},
	})
}

// StatsResponse представляет ответ с данными статистики
type StatsResponse struct {
	TotalPRs     int                        `json:"total_prs"`
	TotalUsers   int                        `json:"total_users"`
	PRsByStatus  map[string]int             `json:"prs_by_status"`
	ActiveUsers  int                        `json:"active_users"`
	TopReviewers []map[string]interface{}   `json:"top_reviewers"`
}
