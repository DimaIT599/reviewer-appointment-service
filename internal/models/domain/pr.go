package domain

import "time"

const (
	PRStatusOpen   = "OPEN"
	PRStatusMerged = "MERGED"
)

type PullRequest struct {
	ID              int64      `json:"id"`
	PullRequestID   string     `json:"pull_request_id"`
	PullRequestName string     `json:"pull_request_name"`
	AuthorID        int64      `json:"author_id"`
	StatusID        int        `json:"status_id"`
	MergedAt        *time.Time `json:"merged_at"`
	CreatedAt       time.Time  `json:"created_at"`
	Author          *User      `json:"author,omitempty"`
	Reviewers       []User     `json:"reviewers,omitempty"`
}
