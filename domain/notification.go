package domain

import (
	"context"
	"time"
)

type AttendanceNotificationHistory struct {
	NotificationHistoryID int       `gorm:"primaryKey;autoIncrement" json:"notification_history_id"`
	SubjectCode           string    `gorm:"not null" json:"subject_code"`
	Subject               Subject   `gorm:"foreignKey:SubjectCode;references:SubjectCode;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"subject"`
	StudentNSN            string    `gorm:"not null" json:"student_nsn"`
	Student               Student   `gorm:"foreignKey:StudentNSN;references:StudentNSN;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"student"`
	ParentID              int       `gorm:"not null;index" json:"parent_id"`
	Parent                Parent    `gorm:"foreignKey:ParentID;references:ParentID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"parent"`
	UserID                int       `gorm:"not null" json:"user_id"`
	User                  User      `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"user"`
	WhatsappStatus        bool      `gorm:"not null" json:"whatsapp"`
	EmailStatus           bool      `gorm:"not null" json:"email"`
	CreatedAt             time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type NotificationRepo interface {
	GetAllAttendanceNotificationHistory(ctx context.Context) (*[]AttendanceNotificationHistoryResponse, error)
}

type NotificationUseCase interface {
	GetAllAttendanceNotificationHistory(ctx context.Context) (*[]AttendanceNotificationHistoryResponse, error)
}
