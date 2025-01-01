package models

import "time"

type Challenge struct {
	ID           int               `json:"id"`
	CreatorID    int               `json:"creatorId"`
	OpponentID   int               `json:"opponentId"`
	StartDate    time.Time         `json:"startDate"`
	EndDate      time.Time         `json:"endDate"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"createdAt"`
	CreatorName  string            `json:"creatorName,omitempty"`
	OpponentName string            `json:"opponentName,omitempty"`
	Results      []ChallengeResult `json:"results,omitempty"`
}

type ChallengeResult struct {
	ChallengeID   int     `json:"challengeId"`
	UserID        int     `json:"userId"`
	Username      string  `json:"username"`
	InitialWeight float64 `json:"initialWeight"`
	FinalWeight   float64 `json:"finalWeight"`
	Progress      float64 `json:"progress"`
}
