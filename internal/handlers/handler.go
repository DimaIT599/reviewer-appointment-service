package handlers

import (
	"reviewer-appointment-service/internal/errors"
	"reviewer-appointment-service/internal/services"
	"reviewer-appointment-service/internal/storage"
)

type Handler struct {
	userService *services.UserService
	teamService *services.TeamService
	prService   *services.PRService
	statsRepo   storage.StatsRepository
}

func NewHandler(userService *services.UserService, teamService *services.TeamService, prService *services.PRService, statsRepo storage.StatsRepository) *Handler {
	return &Handler{
		userService: userService,
		teamService: teamService,
		prService:   prService,
		statsRepo:   statsRepo,
	}
}

type Response struct {
	Data  interface{}    `json:"data,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func errorResponse(err error) (int, *Response) {
	switch e := err.(type) {
	case *errors.AppError:
		return getHTTPStatus(e.Code), &Response{
			Error: &ErrorResponse{
				Code:    string(e.Code),
				Message: e.Message,
			},
		}
	default:
		return 500, &Response{
			Error: &ErrorResponse{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			},
		}
	}
}

func getHTTPStatus(code errors.ErrorCode) int {
	switch code {
	case errors.NotFound:
		return 404
	case errors.TeamExists, errors.PRExists:
		return 400
	case errors.PRMerged, errors.NotAssigned, errors.NoCandidate:
		return 409
	default:
		return 500
	}
}
