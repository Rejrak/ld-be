package models

import (
	"time"

	"github.com/google/uuid"
)

type TrainingPlan struct {
	CustomModel
	Name        string    `gorm:"column:name;not null" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	StartDate   time.Time `gorm:"column:start_date" json:"startDate"`
	EndDate     time.Time `gorm:"column:end_date" json:"endDate"`

	UserID uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	User   User      `gorm:"foreignKey:UserID" json:"user"`

	Workouts []Workout `gorm:"foreignKey:TrainingPlanID" json:"workouts"`
}

type Workout struct {
	CustomModel

	Name string

	TrainingPlanID uuid.UUID    `gorm:"type:uuid;not null" json:"traningPlanId"`
	TrainingPlan   TrainingPlan `gorm:"foreignKey:TrainingPlanID" json:"TrainingPlan"`

	Exercises []Exercise `gorm:"foreignKey:WorkoutID" json:"exerciseID"`
}

type Exercise struct {
	CustomModel

	Name string

	WorkoutID uuid.UUID `gorm:"type:uuid;not null" json:"workoutID"`
	Workout   Workout   `gorm:"foreignKey:WorkoutID" json:"TrainingPlan"`
}
