package domain

import "time"

type SafeStaffData struct {
	UserID    int        `json:"user_id"`
	Username  string     `json:"username"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type SafeStaffUpdatePayload struct {
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Role      string    `json:"role"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StuAndPar struct {
	Student StudentBase `json:"student"`
	Parent  ParentBase  `json:"parent"`
}

type StudentBase struct {
	StudentID int    `gorm:"primaryKey;autoIncrement" json:"student_id"`
	Name      string `gorm:"type:varchar(150);not null;unique" json:"name" valid:"required~Name is required"`
	Class     string `gorm:"type:varchar(3);not null" json:"class" valid:"required~Class is required"`
	Gender    string `gorm:"type:gender_enum;not null" json:"gender" valid:"required~Gender is required"`
	Telephone string `gorm:"type:varchar(15);not null" json:"telephone" valid:"required~Telephone is required"`
	ParentID  int    `gorm:"not null" json:"parent_id"`
}

type ParentBase struct {
	ParentID  int     `gorm:"primaryKey;autoIncrement" json:"parent_id"`
	Name      string  `gorm:"type:varchar(150);not null;unique" json:"name" valid:"required~Name is required"`
	Gender    string  `gorm:"type:gender_enum;not null" json:"gender" valid:"required~Gender is required"`
	Telephone string  `gorm:"type:varchar(15);not null" json:"telephone" valid:"required~Telephone is required"`
	Email     *string `gorm:"type:varchar(255)" json:"email" valid:"email~Invalid email format"`
}
