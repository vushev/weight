package models

import "time"

type WeightRecord struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	Weight    float64   `json:"weight"`
	CreatedAt time.Time `json:"createdAt"`
}

type WeightRecordInput struct {
	Weight    float64 `json:"weight"`
	CreatedAt string  `json:"createdAt"`
}

type WeightStats struct {
	CurrentWeight  float64        `json:"currentWeight"`
	InitialWeight  float64        `json:"initialWeight"`
	PreviousWeight float64        `json:"previousWeight"`
	TotalProgress  float64        `json:"totalProgress"`
	DailyProgress  float64        `json:"dailyProgress"`
	BMI            float64        `json:"bmi"`
	Height         float64        `json:"height"`
	History        []WeightRecord `json:"history"`
}

func CalculateProgress(initialWeight, currentWeight float64) float64 {
	if initialWeight == 0 {
		return 0
	}
	return ((initialWeight - currentWeight) / initialWeight) * 100
}

func CalculateBMI(weight, height float64) float64 {
	if height == 0 {
		return 0
	}
	heightInMeters := height / 100
	return weight / (heightInMeters * heightInMeters)
}
