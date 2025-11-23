package handlers

import (
	"net/http"
	"strconv"

	"reviewer-appointment-service/internal/models/domain"

	"github.com/gin-gonic/gin"
)

// CreateTeam создает команду с участниками
// @Summary Создать команду с участниками
// @Description Создает команду и обновляет/создает пользователей
// @Tags Teams
// @Accept json
// @Produce json
// @Param team body domain.Team true "Данные команды"
// @Success 201 {object} Response{data=domain.Team}
// @Failure 400 {object} Response
// @Failure 500 {object} Response
// @Router /team/add [post]
func (h *Handler) CreateTeam(c *gin.Context) {
	var team domain.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	createdTeam, err := h.teamService.CreateTeam(c.Request.Context(), &team)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusCreated, map[string]interface{}{
		"team": createdTeam,
	})
}

// GetTeam возвращает команду с участниками
// @Summary Получить команду с участниками
// @Description Возвращает команду и список её участников
// @Tags Teams
// @Produce json
// @Param team_name query string true "Название команды"
// @Success 200 {object} Response{data=domain.Team}
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /team/get [get]
func (h *Handler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "MISSING_PARAM",
				Message: "team_name is required",
			},
		})
		return
	}

	team, err := h.teamService.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, team)
}

// DeactivateTeamUsers массово деактивирует пользователей команды
// @Summary Массовая деактивация пользователей команды
// @Description Деактивирует всех пользователей указанной команды
// @Tags Teams
// @Accept json
// @Produce json
// @Param team_id path string true "ID команды"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /teams/{team_id}/deactivate [post]
func (h *Handler) DeactivateTeamUsers(c *gin.Context) {
	teamID := c.Param("team_id")
	if teamID == "" {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "MISSING_PARAM",
				Message: "team_id is required",
			},
		})
		return
	}
	ID, err := strconv.Atoi(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Error: &ErrorResponse{
				Code:    "INVALID_PARAM",
				Message: "team_id is not a valid integer",
			},
		})
		return
	}

	err = h.teamService.DeactivateTeamUsers(c.Request.Context(), int64(ID))
	if err != nil {
		status, resp := errorResponse(err)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, Response{
		Data: map[string]string{"status": "users deactivated"},
	})
}
