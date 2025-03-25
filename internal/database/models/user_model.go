package models

import "github.com/google/uuid"

type User struct {
	CustomModel

	KCID      uuid.UUID `gorm:"type:uuid;unique;column:kc_id" json:"kcId"`
	FirstName string    `gorm:"column:firstname;not null" json:"firstName"`
	LastName  string    `gorm:"column:lastname;not null" json:"lastName"`
	Nickname  string    `json:"nickname"`
	Admin     bool      `gorm:"column:admin;default:false" json:"admin"`

	TrainingPlans []TrainingPlan `gorm:"foreignKey:UserID" json:"trainingPlans"`
}
