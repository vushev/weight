package models

import (
	"database/sql"
	"fmt"
	"time"
	"log"
)

// CalorieSettings съдържа настройките за калории на потребителя
type CalorieSettings struct {
	UserID        int       `json:"userId"`
	Gender        string    `json:"gender"`
	Age           int       `json:"age"`
	ActivityLevel string    `json:"activityLevel"`
	Goal          string    `json:"goal"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// CalorieCalculation съдържа изчислените калории
type CalorieCalculation struct {
	BMR             float64 `json:"bmr"`
	MaintenanceKcal float64 `json:"maintenanceKcal"`
	TargetKcal      float64 `json:"targetKcal"`
}

// DailyCalorieLog представлява дневен лог за калории
type DailyCalorieLog struct {
	ID           int       `json:"id"`
	UserID       int       `json:"userId"`
	Date         time.Time `json:"date"`
	FoodKcal     float64   `json:"foodKcal"`
	ActivityKcal float64   `json:"activityKcal"`
	Notes        sql.NullString    `json:"notes"`
}

type Date struct {
    time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	log.Printf("Date: %v", d.Time)
    return []byte(fmt.Sprintf(`"%s"`, d.Format("2006-01-02"))), nil
}
func Now() Date {
    return Date{Time: time.Now()}
}

// FoodEntry представлява запис за храна
type FoodEntry struct {
	ID         int       `json:"id"`
	UserID     int       `json:"userId"`
	LogID      int       `json:"logId"`
	Name       string    `json:"name"`
	MealTypeID int       `json:"mealTypeId"`
	MealType   string    `json:"mealType,omitempty"` // За JSON отговори
	Calories   float64   `json:"calories"`
	Protein    float64   `json:"protein"`
	Carbs      float64   `json:"carbs"`
	Fat        float64   `json:"fat"`
	Notes      string    `json:"notes"`
	Time       time.Time `json:"time" default:"now" time_format:"2006-01-02"`
}

// ActivityEntry представлява запис за физическа активност
type ActivityEntry struct {
	ID       int       `json:"id"`
	UserID   int       `json:"userId"`
	LogID    int       `json:"logId"`
	Type     string    `json:"type"`
	Duration int       `json:"duration"`
	Calories float64   `json:"calories"`
	Time     time.Time `json:"time" time_format:"2006-01-02"`
}

// CalorieIntake представлява входящи данни за калории
type CalorieIntake struct {
	Date     time.Time `json:"date"`
	Calories float64   `json:"calories"`
	Type     string    `json:"type"` // "food" или "activity"
} 

type MealType struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt" default:"now"`
	UpdatedAt time.Time `json:"updatedAt" default:"now"`
}