package errors

import "fmt"

type ErrorCode string

const (
	TeamExists  ErrorCode = "TEAM_EXISTS"
	PRExists    ErrorCode = "PR_EXISTS"
	PRMerged    ErrorCode = "PR_MERGED"
	NotAssigned ErrorCode = "NOT_ASSIGNED"
	NoCandidate ErrorCode = "NO_CANDIDATE"
	NotFound    ErrorCode = "NOT_FOUND"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

var (
	ErrTeamExists  = NewAppError(TeamExists, "team_name already exists")
	ErrPRExists    = NewAppError(PRExists, "PR id already exists")
	ErrPRMerged    = NewAppError(PRMerged, "cannot modify merged PR")
	ErrNotAssigned = NewAppError(NotAssigned, "reviewer is not assigned to this PR")
	ErrNoCandidate = NewAppError(NoCandidate, "no active replacement candidate in team")
	ErrNotFound    = NewAppError(NotFound, "resource not found")
)
