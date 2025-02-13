package domain

import (
	"time"
)

type Parent struct {
	ParentID  int       `gorm:"primaryKey;autoIncrement" json:"parent_id"`
	Name      string    `gorm:"type:varchar(150);not null;" json:"name" valid:"required~Name is required"`
	Gender    string    `gorm:"type:gender_enum;not null" json:"gender" valid:"required~Gender is required,in(male|female)~Invalid gender"`
	Telephone string    `gorm:"type:varchar(13);not null;" json:"telephone" valid:"required~Telephone is required"`
	Email     *string   `gorm:"type:varchar(255)" json:"email" valid:"email~Invalid email format,optional"` // Make email optional
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
