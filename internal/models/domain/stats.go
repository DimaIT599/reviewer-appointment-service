package domain

type ReviewerStats struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	ReviewCount int    `json:"review_count"`
}
